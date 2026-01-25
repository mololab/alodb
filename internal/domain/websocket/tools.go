package websocket

// SchemaQuery defines a query to be executed by the client for schema extraction
type SchemaQuery struct {
	ID          string
	Name        string
	Description string
	Query       string
	PerTable    bool
}

// Schema extraction queries - sent to client one by one
var SchemaQueries = struct {
	GetDatabaseName SchemaQuery
	GetTables       SchemaQuery
	GetColumns      SchemaQuery
	GetPrimaryKey   SchemaQuery
	GetForeignKeys  SchemaQuery
	GetIndexes      SchemaQuery
}{
	GetDatabaseName: SchemaQuery{
		ID:          "get_database_name",
		Name:        "Getting database name",
		Description: "Retrieves the current database name",
		Query:       "SELECT current_database()",
		PerTable:    false,
	},
	GetTables: SchemaQuery{
		ID:          "get_tables",
		Name:        "Discovering tables",
		Description: "Lists all tables in the public schema",
		Query: `SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_type = 'BASE TABLE'
			ORDER BY table_name`,
		PerTable: false,
	},
	GetColumns: SchemaQuery{
		ID:          "get_columns",
		Name:        "Reading columns for %s",
		Description: "Gets column definitions for table %s",
		Query: `SELECT 
				column_name,
				data_type,
				is_nullable,
				COALESCE(column_default, '') as column_default,
				COALESCE(col_description((table_schema || '.' || table_name)::regclass::oid, ordinal_position), '') as column_comment
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = '%s'
			ORDER BY ordinal_position`,
		PerTable: true,
	},
	GetPrimaryKey: SchemaQuery{
		ID:          "get_primary_key",
		Name:        "Finding primary key for %s",
		Description: "Gets primary key columns for table %s",
		Query: `SELECT a.attname
			FROM pg_index i
			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
			WHERE i.indrelid = ('public.' || '%s')::regclass
			AND i.indisprimary
			ORDER BY array_position(i.indkey, a.attnum)`,
		PerTable: true,
	},
	GetForeignKeys: SchemaQuery{
		ID:          "get_foreign_keys",
		Name:        "Finding foreign keys for %s",
		Description: "Gets foreign key relationships for table %s",
		Query: `SELECT
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
			AND tc.table_name = '%s'
			ORDER BY tc.constraint_name, kcu.ordinal_position`,
		PerTable: true,
	},
	GetIndexes: SchemaQuery{
		ID:          "get_indexes",
		Name:        "Reading indexes for %s",
		Description: "Gets index definitions for table %s",
		Query: `SELECT
				i.relname as index_name,
				array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)) as column_names,
				ix.indisunique as is_unique
			FROM pg_index ix
			JOIN pg_class i ON ix.indexrelid = i.oid
			JOIN pg_class t ON ix.indrelid = t.oid
			JOIN pg_namespace n ON t.relnamespace = n.oid
			JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
			WHERE t.relname = '%s'
			AND n.nspname = 'public'
			AND NOT ix.indisprimary
			GROUP BY i.relname, ix.indisunique
			ORDER BY i.relname`,
		PerTable: true,
	},
}
