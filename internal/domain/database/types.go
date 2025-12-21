package database

// TableSchema represents a database table structure
type TableSchema struct {
	Name        string         `json:"name"`
	Columns     []ColumnSchema `json:"columns"`
	PrimaryKey  []string       `json:"primary_key,omitempty"`
	ForeignKeys []ForeignKey   `json:"foreign_keys,omitempty"`
	Indexes     []IndexSchema  `json:"indexes,omitempty"`
}

// ColumnSchema represents a column in a table
type ColumnSchema struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	IsNullable bool   `json:"is_nullable"`
	Default    string `json:"default,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	Name             string   `json:"name"`
	Columns          []string `json:"columns"`
	ReferencedTable  string   `json:"referenced_table"`
	ReferencedColumn []string `json:"referenced_columns"`
}

// IndexSchema represents an index on a table
type IndexSchema struct {
	Name     string   `json:"name"`
	Columns  []string `json:"columns"`
	IsUnique bool     `json:"is_unique"`
}

// DatabaseSchema represents the complete database schema
type DatabaseSchema struct {
	DatabaseName string        `json:"database_name"`
	Tables       []TableSchema `json:"tables"`
}

// QueryResult represents the result of a query generation
type QueryResult struct {
	Query       string `json:"query"`
	Explanation string `json:"explanation"`
	IsReadOnly  bool   `json:"is_read_only"`
}
