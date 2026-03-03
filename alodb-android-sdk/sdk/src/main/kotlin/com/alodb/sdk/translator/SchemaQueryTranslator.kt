package com.alodb.sdk.translator

import com.alodb.sdk.database.DatabaseDriver

/**
 * Translates PostgreSQL information_schema / pg_catalog queries sent by
 * the AloDB server into SQLite equivalents, then maps result columns to
 * the names the server expects.
 *
 * The server sends exactly 6 query types (defined in tools.go). This
 * translator pattern-matches each one and produces the correct SQLite
 * output.
 */
object SchemaQueryTranslator {

    data class TranslatedQuery(
        val sql: String,
        val postProcess: ((List<Map<String, Any?>>) -> List<Map<String, Any?>>)? = null,
    )

    private val TABLE_NAME_IN_QUOTES = Regex("""['"](\w+)['"]""")
    private val TABLE_NAME_RELNAME = Regex("""relname\s*=\s*['"](\w+)['"]""")
    private val TABLE_NAME_PARAM = Regex("""table_name\s*=\s*['"](\w+)['"]""")

    fun translate(pgQuery: String, driver: DatabaseDriver? = null): TranslatedQuery {
        val sql = pgQuery.trim()

        return when {
            sql.contains("current_database", ignoreCase = true) ->
                translateCurrentDatabase(driver)

            sql.contains("information_schema.tables", ignoreCase = true) ->
                translateGetTables()

            sql.contains("information_schema.columns", ignoreCase = true) ->
                translateGetColumns(sql)

            // Primary key: indisprimary WITHOUT "NOT" preceding it
            sql.contains("indisprimary", ignoreCase = true)
                    && !sql.contains("NOT ix.indisprimary", ignoreCase = true)
                    && !sql.contains("NOT indisprimary", ignoreCase = true) ->
                translateGetPrimaryKey(sql)

            sql.contains("FOREIGN KEY", ignoreCase = true) ->
                translateGetForeignKeys(sql)

            // Indexes: has "NOT ix.indisprimary" or "NOT indisprimary"
            sql.contains("NOT ix.indisprimary", ignoreCase = true)
                    || sql.contains("NOT indisprimary", ignoreCase = true) ->
                translateGetIndexes(sql, driver)

            else -> TranslatedQuery(sql)
        }
    }

    private fun translateCurrentDatabase(driver: DatabaseDriver?): TranslatedQuery {
        val dbName = driver?.getDatabaseName() ?: "main"
        return TranslatedQuery(
            sql = "SELECT '$dbName' AS current_database",
        )
    }

    private fun translateGetTables(): TranslatedQuery {
        return TranslatedQuery(
            sql = """
                SELECT name AS table_name 
                FROM sqlite_master 
                WHERE type = 'table' 
                AND name NOT LIKE 'sqlite_%' 
                AND name NOT LIKE 'android_%'
                AND name NOT LIKE 'room_%'
                ORDER BY name
            """.trimIndent(),
        )
    }

    private fun translateGetColumns(pgQuery: String): TranslatedQuery {
        val tableName = extractTableName(pgQuery)
        return TranslatedQuery(
            sql = "PRAGMA table_info('$tableName')",
            postProcess = { rows ->
                rows.map { row ->
                    mapOf(
                        "column_name" to (row["name"] ?: ""),
                        "data_type" to mapSqliteType(row["type"]?.toString() ?: ""),
                        "is_nullable" to if (row["notnull"]?.toString() == "0") "YES" else "NO",
                        "column_default" to (row["dflt_value"]?.toString() ?: ""),
                        "column_comment" to "",
                    )
                }
            },
        )
    }

    private fun translateGetPrimaryKey(pgQuery: String): TranslatedQuery {
        val tableName = extractTableName(pgQuery)
        return TranslatedQuery(
            sql = "PRAGMA table_info('$tableName')",
            postProcess = { rows ->
                rows.filter { row ->
                    val pk = row["pk"]
                    pk != null && pk.toString() != "0"
                }.map { row ->
                    mapOf("attname" to (row["name"] ?: ""))
                }
            },
        )
    }

    private fun translateGetForeignKeys(pgQuery: String): TranslatedQuery {
        val tableName = extractTableName(pgQuery)
        return TranslatedQuery(
            sql = "PRAGMA foreign_key_list('$tableName')",
            postProcess = { rows ->
                rows.map { row ->
                    val id = row["id"]?.toString() ?: "0"
                    mapOf(
                        "constraint_name" to "fk_${tableName}_$id",
                        "column_name" to (row["from"] ?: ""),
                        "foreign_table_name" to (row["table"] ?: ""),
                        "foreign_column_name" to (row["to"] ?: ""),
                    )
                }
            },
        )
    }

    private fun translateGetIndexes(pgQuery: String, driver: DatabaseDriver?): TranslatedQuery {
        val tableName = extractTableName(pgQuery)
        return TranslatedQuery(
            sql = "PRAGMA index_list('$tableName')",
            postProcess = { indexRows ->
                indexRows.mapNotNull { indexRow ->
                    val indexName = indexRow["name"]?.toString() ?: return@mapNotNull null
                    val isUnique = indexRow["unique"]?.toString() == "1"

                    // Skip auto-generated primary key indexes
                    if (indexName.startsWith("sqlite_autoindex")) return@mapNotNull null

                    val columns = if (driver != null) {
                        try {
                            driver.execute("PRAGMA index_info('$indexName')").map { col ->
                                col["name"]?.toString() ?: ""
                            }
                        } catch (_: Exception) {
                            emptyList()
                        }
                    } else {
                        emptyList()
                    }

                    mapOf(
                        "index_name" to indexName,
                        "column_names" to columns.joinToString(",", "{", "}"),
                        "is_unique" to isUnique,
                    )
                }
            },
        )
    }

    private fun extractTableName(query: String): String {
        TABLE_NAME_PARAM.find(query)?.groupValues?.get(1)?.let { return it }
        TABLE_NAME_RELNAME.find(query)?.groupValues?.get(1)?.let { return it }

        val allMatches = TABLE_NAME_IN_QUOTES.findAll(query).map { it.groupValues[1] }.toList()
        val excluded = setOf(
            "public", "BASE", "TABLE", "YES", "NO",
            "FOREIGN", "KEY", "PRIMARY",
        )
        return allMatches.lastOrNull { it !in excluded } ?: run {
            android.util.Log.w("SchemaQueryTranslator", "Could not extract table name from query: ${query.take(120)}")
            "unknown"
        }
    }

    private fun mapSqliteType(sqliteType: String): String {
        val upper = sqliteType.uppercase()
        return when {
            upper.contains("INT") -> "integer"
            upper.contains("CHAR") || upper.contains("CLOB") || upper.contains("TEXT") -> "text"
            upper.contains("BLOB") || upper.isEmpty() -> "blob"
            upper.contains("REAL") || upper.contains("FLOA") || upper.contains("DOUB") -> "real"
            upper.contains("BOOL") -> "boolean"
            upper.contains("DATE") || upper.contains("TIME") -> "timestamp without time zone"
            upper.contains("NUMERIC") || upper.contains("DECIMAL") -> "numeric"
            else -> sqliteType.lowercase()
        }
    }
}
