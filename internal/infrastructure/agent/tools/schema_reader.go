package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"github.com/mololab/alodb/internal/domain/database"
	"github.com/mololab/alodb/pkg/logger"
)

// SchemaReaderInput represents the input for the schema reader tool
type SchemaReaderInput struct{}

// SchemaReaderOutput represents the output from the schema reader tool
type SchemaReaderOutput struct {
	Status  string                   `json:"status"`
	Schema  *database.DatabaseSchema `json:"schema,omitempty"`
	Message string                   `json:"message,omitempty"`
}

// ReadSchemaFromDatabase reads the database schema directly from PostgreSQL
func ReadSchemaFromDatabase(connectionString string) (SchemaReaderOutput, error) {
	if connectionString == "" {
		return SchemaReaderOutput{
			Status:  "error",
			Message: "No database connection configured. Please provide a connection string.",
		}, nil
	}

	ctx := context.Background()

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Error().Err(err).Msg("failed to open database")
		return SchemaReaderOutput{
			Status:  "error",
			Message: fmt.Sprintf("failed to connect to database: %v", err),
		}, nil
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		logger.Error().Err(err).Msg("failed to ping database")
		return SchemaReaderOutput{
			Status:  "error",
			Message: fmt.Sprintf("failed to ping database: %v", err),
		}, nil
	}

	schema, err := extractPostgresSchema(ctx, db)
	if err != nil {
		logger.Error().Err(err).Msg("failed to extract schema")
		return SchemaReaderOutput{
			Status:  "error",
			Message: fmt.Sprintf("failed to extract schema: %v", err),
		}, nil
	}

	logger.Info().Int("tables", len(schema.Tables)).Msg("schema extracted")
	return SchemaReaderOutput{
		Status: "success",
		Schema: schema,
	}, nil
}

// GetSchemaAsJSON returns the schema as a formatted JSON string for LLM consumption
func GetSchemaAsJSON(schema *database.DatabaseSchema) (string, error) {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// extractPostgresSchema extracts the complete schema from a PostgreSQL database
func extractPostgresSchema(ctx context.Context, db *sql.DB) (*database.DatabaseSchema, error) {
	schema := &database.DatabaseSchema{}

	var dbName string
	err := db.QueryRowContext(ctx, "SELECT current_database()").Scan(&dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database name: %w", err)
	}
	schema.DatabaseName = dbName

	tables, err := getTables(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	for _, tableName := range tables {
		tableSchema, err := getTableSchema(ctx, db, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get schema for table %s: %w", tableName, err)
		}
		schema.Tables = append(schema.Tables, *tableSchema)
	}

	return schema, nil
}

// getTables returns all user tables in the database
func getTables(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// getTableSchema returns the schema for a specific table
func getTableSchema(ctx context.Context, db *sql.DB, tableName string) (*database.TableSchema, error) {
	tableSchema := &database.TableSchema{
		Name: tableName,
	}

	// columns
	columns, err := getColumns(ctx, db, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	tableSchema.Columns = columns

	// primary key
	primaryKey, err := getPrimaryKey(ctx, db, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary key: %w", err)
	}
	tableSchema.PrimaryKey = primaryKey

	// foreign keys
	foreignKeys, err := getForeignKeys(ctx, db, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys: %w", err)
	}
	tableSchema.ForeignKeys = foreignKeys

	// indexes
	indexes, err := getIndexes(ctx, db, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	tableSchema.Indexes = indexes

	return tableSchema, nil
}

// getColumns returns all columns for a table
func getColumns(ctx context.Context, db *sql.DB, tableName string) ([]database.ColumnSchema, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			COALESCE(column_default, '') as column_default,
			COALESCE(col_description((table_schema || '.' || table_name)::regclass::oid, ordinal_position), '') as column_comment
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []database.ColumnSchema
	for rows.Next() {
		var col database.ColumnSchema
		var isNullable string
		if err := rows.Scan(&col.Name, &col.DataType, &isNullable, &col.Default, &col.Comment); err != nil {
			return nil, err
		}
		col.IsNullable = isNullable == "YES"
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// getPrimaryKey returns the primary key columns for a table
func getPrimaryKey(ctx context.Context, db *sql.DB, tableName string) ([]string, error) {
	query := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = ('public.' || $1)::regclass
		AND i.indisprimary
		ORDER BY array_position(i.indkey, a.attnum)
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, err
		}
		columns = append(columns, colName)
	}

	return columns, rows.Err()
}

// getForeignKeys returns foreign key constraints for a table
func getForeignKeys(ctx context.Context, db *sql.DB, tableName string) ([]database.ForeignKey, error) {
	query := `
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_schema = 'public'
		AND tc.table_name = $1
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fkMap := make(map[string]*database.ForeignKey)
	for rows.Next() {
		var constraintName, colName, refTable, refCol string
		if err := rows.Scan(&constraintName, &colName, &refTable, &refCol); err != nil {
			return nil, err
		}

		if fk, ok := fkMap[constraintName]; ok {
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

	var foreignKeys []database.ForeignKey
	for _, fk := range fkMap {
		foreignKeys = append(foreignKeys, *fk)
	}

	return foreignKeys, rows.Err()
}

// getIndexes returns all indexes for a table
func getIndexes(ctx context.Context, db *sql.DB, tableName string) ([]database.IndexSchema, error) {
	query := `
		SELECT
			i.relname as index_name,
			array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)) as column_names,
			ix.indisunique as is_unique
		FROM pg_index ix
		JOIN pg_class i ON ix.indexrelid = i.oid
		JOIN pg_class t ON ix.indrelid = t.oid
		JOIN pg_namespace n ON t.relnamespace = n.oid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relname = $1
		AND n.nspname = 'public'
		AND NOT ix.indisprimary
		GROUP BY i.relname, ix.indisunique
		ORDER BY i.relname
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []database.IndexSchema
	for rows.Next() {
		var idx database.IndexSchema
		var columns []string
		if err := rows.Scan(&idx.Name, pq.Array(&columns), &idx.IsUnique); err != nil {
			return nil, err
		}
		idx.Columns = columns
		indexes = append(indexes, idx)
	}

	return indexes, rows.Err()
}
