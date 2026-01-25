package schema

import (
	"fmt"
	"strings"
	"time"

	"github.com/mololab/alodb/internal/domain/database"
	ws "github.com/mololab/alodb/internal/domain/websocket"
	"github.com/mololab/alodb/pkg/logger"
)

const queryTimeout = 30 * time.Second

type QueryExecutor interface {
	ExecuteQuery(name, description, query string, step, totalSteps int, timeout time.Duration) (*ws.QueryResultPayload, error)
}

type Reader struct {
	executor QueryExecutor
}

func NewReader(executor QueryExecutor) *Reader {
	return &Reader{executor: executor}
}

// ReadSchema extracts the complete database schema by sending queries to client
func (r *Reader) ReadSchema() (*database.DatabaseSchema, error) {
	schema := &database.DatabaseSchema{}

	// Step 1: get database name
	dbName, err := r.getDatabaseName()
	if err != nil {
		return nil, fmt.Errorf("failed to get database name: %w", err)
	}
	schema.DatabaseName = dbName

	// Step 2: get all tables
	tables, err := r.getTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	if len(tables) == 0 {
		logger.Info().Msg("no tables found in database")
		return schema, nil
	}

	// calculate total steps for progress reporting
	// 2 base queries + 4 queries per table (columns, pk, fk, indexes)
	totalSteps := 2 + (len(tables) * 4)
	currentStep := 2

	// Step 3: get schema for each table
	for _, tableName := range tables {
		tableSchema, err := r.getTableSchema(tableName, &currentStep, totalSteps)
		if err != nil {
			return nil, fmt.Errorf("failed to get schema for table %s: %w", tableName, err)
		}
		schema.Tables = append(schema.Tables, *tableSchema)
	}

	logger.Info().Int("tables", len(schema.Tables)).Msg("schema extraction complete")
	return schema, nil
}

func (r *Reader) getDatabaseName() (string, error) {
	q := ws.SchemaQueries.GetDatabaseName

	result, err := r.executor.ExecuteQuery(q.Name, q.Description, q.Query, 1, 2, queryTimeout)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf(result.Error)
	}

	// extract database name from result rows
	if len(result.Rows) > 0 {
		if row, ok := result.Rows[0].(map[string]any); ok {
			if name, ok := row["current_database"].(string); ok {
				return name, nil
			}
		}
		// handle array format
		if row, ok := result.Rows[0].([]any); ok && len(row) > 0 {
			if name, ok := row[0].(string); ok {
				return name, nil
			}
		}
	}

	return "unknown", nil
}

func (r *Reader) getTables() ([]string, error) {
	q := ws.SchemaQueries.GetTables

	result, err := r.executor.ExecuteQuery(q.Name, q.Description, q.Query, 2, 2, queryTimeout)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	var tables []string
	for _, row := range result.Rows {
		if rowMap, ok := row.(map[string]any); ok {
			if name, ok := rowMap["table_name"].(string); ok {
				tables = append(tables, name)
			}
		}
		// handle array format
		if rowArr, ok := row.([]any); ok && len(rowArr) > 0 {
			if name, ok := rowArr[0].(string); ok {
				tables = append(tables, name)
			}
		}
	}

	return tables, nil
}

func (r *Reader) getTableSchema(tableName string, currentStep *int, totalSteps int) (*database.TableSchema, error) {
	tableSchema := &database.TableSchema{Name: tableName}

	// get columns
	columns, err := r.getColumns(tableName, currentStep, totalSteps)
	if err != nil {
		return nil, err
	}
	tableSchema.Columns = columns

	// get primary key
	pk, err := r.getPrimaryKey(tableName, currentStep, totalSteps)
	if err != nil {
		return nil, err
	}
	tableSchema.PrimaryKey = pk

	// get foreign keys
	fks, err := r.getForeignKeys(tableName, currentStep, totalSteps)
	if err != nil {
		return nil, err
	}
	tableSchema.ForeignKeys = fks

	// get indexes
	indexes, err := r.getIndexes(tableName, currentStep, totalSteps)
	if err != nil {
		return nil, err
	}
	tableSchema.Indexes = indexes

	return tableSchema, nil
}

func (r *Reader) getColumns(tableName string, currentStep *int, totalSteps int) ([]database.ColumnSchema, error) {
	q := ws.SchemaQueries.GetColumns
	*currentStep++

	name := fmt.Sprintf(q.Name, tableName)
	desc := fmt.Sprintf(q.Description, tableName)
	query := fmt.Sprintf(q.Query, tableName)

	result, err := r.executor.ExecuteQuery(name, desc, query, *currentStep, totalSteps, queryTimeout)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	var columns []database.ColumnSchema
	for _, row := range result.Rows {
		col := database.ColumnSchema{}

		if rowMap, ok := row.(map[string]any); ok {
			col.Name, _ = rowMap["column_name"].(string)
			col.DataType, _ = rowMap["data_type"].(string)
			isNullable, _ := rowMap["is_nullable"].(string)
			col.IsNullable = isNullable == "YES"
			col.Default, _ = rowMap["column_default"].(string)
			col.Comment, _ = rowMap["column_comment"].(string)
		}

		if col.Name != "" {
			columns = append(columns, col)
		}
	}

	return columns, nil
}

func (r *Reader) getPrimaryKey(tableName string, currentStep *int, totalSteps int) ([]string, error) {
	q := ws.SchemaQueries.GetPrimaryKey
	*currentStep++

	name := fmt.Sprintf(q.Name, tableName)
	desc := fmt.Sprintf(q.Description, tableName)
	query := fmt.Sprintf(q.Query, tableName)

	result, err := r.executor.ExecuteQuery(name, desc, query, *currentStep, totalSteps, queryTimeout)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	var columns []string
	for _, row := range result.Rows {
		if rowMap, ok := row.(map[string]any); ok {
			if name, ok := rowMap["attname"].(string); ok {
				columns = append(columns, name)
			}
		}
		if rowArr, ok := row.([]any); ok && len(rowArr) > 0 {
			if name, ok := rowArr[0].(string); ok {
				columns = append(columns, name)
			}
		}
	}

	return columns, nil
}

func (r *Reader) getForeignKeys(tableName string, currentStep *int, totalSteps int) ([]database.ForeignKey, error) {
	q := ws.SchemaQueries.GetForeignKeys
	*currentStep++

	name := fmt.Sprintf(q.Name, tableName)
	desc := fmt.Sprintf(q.Description, tableName)
	query := fmt.Sprintf(q.Query, tableName)

	result, err := r.executor.ExecuteQuery(name, desc, query, *currentStep, totalSteps, queryTimeout)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	fkMap := make(map[string]*database.ForeignKey)
	for _, row := range result.Rows {
		if rowMap, ok := row.(map[string]any); ok {
			constraintName, _ := rowMap["constraint_name"].(string)
			colName, _ := rowMap["column_name"].(string)
			refTable, _ := rowMap["foreign_table_name"].(string)
			refCol, _ := rowMap["foreign_column_name"].(string)

			if constraintName == "" {
				continue
			}

			if fk, exists := fkMap[constraintName]; exists {
				fk.Columns = append(fk.Columns, colName)
				fk.ReferencedColumn = append(fk.ReferencedColumn, refCol)
			} else {
				fkMap[constraintName] = &database.ForeignKey{
					Name:             constraintName,
					Columns:          []string{colName},
					ReferencedTable:  refTable,
					ReferencedColumn: []string{refCol},
				}
			}
		}
	}

	var foreignKeys []database.ForeignKey
	for _, fk := range fkMap {
		foreignKeys = append(foreignKeys, *fk)
	}

	return foreignKeys, nil
}

func (r *Reader) getIndexes(tableName string, currentStep *int, totalSteps int) ([]database.IndexSchema, error) {
	q := ws.SchemaQueries.GetIndexes
	*currentStep++

	name := fmt.Sprintf(q.Name, tableName)
	desc := fmt.Sprintf(q.Description, tableName)
	query := fmt.Sprintf(q.Query, tableName)

	result, err := r.executor.ExecuteQuery(name, desc, query, *currentStep, totalSteps, queryTimeout)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	var indexes []database.IndexSchema
	for _, row := range result.Rows {
		if rowMap, ok := row.(map[string]any); ok {
			idx := database.IndexSchema{}
			idx.Name, _ = rowMap["index_name"].(string)
			idx.IsUnique, _ = rowMap["is_unique"].(bool)

			// handle column names (could be array or string)
			if cols, ok := rowMap["column_names"].([]any); ok {
				for _, c := range cols {
					if s, ok := c.(string); ok {
						idx.Columns = append(idx.Columns, s)
					}
				}
			} else if colStr, ok := rowMap["column_names"].(string); ok {
				// parse postgresql array format: {col1,col2}
				colStr = strings.Trim(colStr, "{}")
				if colStr != "" {
					idx.Columns = strings.Split(colStr, ",")
				}
			}

			if idx.Name != "" {
				indexes = append(indexes, idx)
			}
		}
	}

	return indexes, nil
}
