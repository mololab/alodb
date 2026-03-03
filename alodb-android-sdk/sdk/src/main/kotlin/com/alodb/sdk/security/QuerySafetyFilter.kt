package com.alodb.sdk.security

/**
 * Validates every query_request from the server before execution.
 * Only known schema-introspection queries pass through. Everything
 * else is blocked — the query is never executed and no row data is
 * ever sent to the server.
 *
 * Security: each semicolon-delimited statement is checked independently
 * to prevent multi-statement injection bypasses such as:
 *   SELECT ... ; DELETE FROM ...
 * WITH-clause DML (WITH x AS (DELETE ...) SELECT ...) is caught by
 * scanning all tokens in the full SQL before splitting.
 */
object QuerySafetyFilter {

    private val ALLOWED_KEYWORDS = listOf(
        "current_database",
        "information_schema",
        "sqlite_master",
    )

    private val ALLOWED_PRAGMA = Regex(
        """^\s*PRAGMA\s+(table_info|foreign_key_list|index_list|index_info)\s*\(""",
        RegexOption.IGNORE_CASE,
    )

    private val DML_ANYWHERE = Regex(
        """\b(INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|REPLACE)\s""",
        RegexOption.IGNORE_CASE,
    )

    /**
     * Schema-only mode (default): only schema introspection queries pass.
     */
    fun check(sql: String): FilterResult {
        if (DML_ANYWHERE.containsMatchIn(sql)) {
            return FilterResult.Blocked(
                "Non-SELECT statement detected: ${sql.take(80)}"
            )
        }

        val statements = sql.split(";").map { it.trim() }.filter { it.isNotEmpty() }
        for (stmt in statements) {
            val isAllowed = ALLOWED_KEYWORDS.any { stmt.contains(it, ignoreCase = true) }
                    || ALLOWED_PRAGMA.containsMatchIn(stmt)
            if (!isAllowed) {
                return FilterResult.Blocked(
                    "Statement does not match schema whitelist: ${stmt.take(80)}"
                )
            }
        }

        return if (statements.isEmpty()) {
            FilterResult.Blocked("Empty query")
        } else {
            FilterResult.Allowed
        }
    }

    /**
     * SELECT-only mode (sendQueryResults=true): all SELECT/WITH/PRAGMA queries
     * pass through so query results can be sent back to the LLM. DML is still
     * always blocked.
     */
    fun checkSelectOnly(sql: String): FilterResult {
        if (DML_ANYWHERE.containsMatchIn(sql)) {
            return FilterResult.Blocked(
                "Non-SELECT statement detected: ${sql.take(80)}"
            )
        }

        val statements = sql.split(";").map { it.trim() }.filter { it.isNotEmpty() }
        for (stmt in statements) {
            val upper = stmt.trimStart()
            if (!upper.startsWith("SELECT", ignoreCase = true)
                && !upper.startsWith("WITH", ignoreCase = true)
                && !upper.startsWith("PRAGMA", ignoreCase = true)) {
                return FilterResult.Blocked(
                    "Only SELECT queries are allowed: ${stmt.take(80)}"
                )
            }
        }

        return if (statements.isEmpty()) {
            FilterResult.Blocked("Empty query")
        } else {
            FilterResult.Allowed
        }
    }
}

sealed class FilterResult {
    data object Allowed : FilterResult()
    data class Blocked(val reason: String) : FilterResult()
}
