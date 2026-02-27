package schema

import (
	"fmt"
	"strings"

	"github.com/mololab/alodb/internal/domain/database"
)

var pgTypeToMermaid = map[string]string{
	"integer":                  "int",
	"bigint":                   "bigint",
	"smallint":                 "smallint",
	"serial":                   "serial",
	"bigserial":                "bigserial",
	"numeric":                  "decimal",
	"real":                     "float",
	"double precision":         "double",
	"boolean":                  "bool",
	"character varying":        "varchar",
	"character":                "char",
	"text":                     "text",
	"uuid":                     "uuid",
	"json":                     "json",
	"jsonb":                    "jsonb",
	"date":                     "date",
	"timestamp with time zone": "timestamptz",
	"timestamp without time zone": "timestamp",
	"time with time zone":         "timetz",
	"time without time zone":      "time",
	"bytea":                       "bytea",
	"inet":                        "inet",
	"interval":                    "interval",
}

func mapPgType(pgType string) string {
	if mapped, ok := pgTypeToMermaid[pgType]; ok {
		return mapped
	}
	normalized := strings.ToLower(pgType)
	if mapped, ok := pgTypeToMermaid[normalized]; ok {
		return mapped
	}
	return pgType
}

// GenerateMermaid builds a Mermaid erDiagram string from the given schema,
// filtered to only the tables listed in usedTables.
// If usedTables is empty, all tables are included.
func GenerateMermaid(dbSchema *database.DatabaseSchema, usedTables []string) string {
	if dbSchema == nil || len(dbSchema.Tables) == 0 {
		return ""
	}

	tableSet := make(map[string]bool, len(usedTables))
	for _, t := range usedTables {
		tableSet[strings.ToLower(t)] = true
	}
	filterAll := len(tableSet) == 0

	var filtered []database.TableSchema
	for _, t := range dbSchema.Tables {
		if filterAll || tableSet[strings.ToLower(t.Name)] {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	pkIndex := make(map[string]map[string]bool)
	for _, t := range filtered {
		pks := make(map[string]bool, len(t.PrimaryKey))
		for _, pk := range t.PrimaryKey {
			pks[pk] = true
		}
		pkIndex[t.Name] = pks
	}

	fkColumns := make(map[string]map[string]bool)
	for _, t := range filtered {
		cols := make(map[string]bool)
		for _, fk := range t.ForeignKeys {
			for _, c := range fk.Columns {
				cols[c] = true
			}
		}
		fkColumns[t.Name] = cols
	}

	var b strings.Builder
	b.WriteString("erDiagram\n")

	for _, t := range filtered {
		b.WriteString(fmt.Sprintf("    %s {\n", t.Name))
		for _, col := range t.Columns {
			mType := mapPgType(col.DataType)
			marker := ""
			if pkIndex[t.Name][col.Name] {
				marker = " PK"
			} else if fkColumns[t.Name][col.Name] {
				marker = " FK"
			}
			b.WriteString(fmt.Sprintf("        %s %s%s\n", mType, col.Name, marker))
		}
		b.WriteString("    }\n")
	}

	filteredSet := make(map[string]bool, len(filtered))
	for _, t := range filtered {
		filteredSet[t.Name] = true
	}

	for _, t := range filtered {
		for _, fk := range t.ForeignKeys {
			if !filteredSet[fk.ReferencedTable] {
				continue
			}
			colLabel := strings.Join(fk.Columns, ", ")
			b.WriteString(fmt.Sprintf("    %s }o--|| %s : \"%s\"\n", t.Name, fk.ReferencedTable, colLabel))
		}
	}

	return b.String()
}
