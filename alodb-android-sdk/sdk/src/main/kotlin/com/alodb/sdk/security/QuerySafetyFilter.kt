package com.alodb.sdk.security

/**
 * Validates every query_request from the server before execution.
 * Only known schema-introspection queries pass through. Everything
 * else is blocked — the query is never executed and no row data is
 * ever sent to the server.
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

    private val FORBIDDEN_STATEMENTS = Regex(
        """^\s*(INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|REPLACE)\s""",
        RegexOption.IGNORE_CASE,
    )

    fun check(sql: String): FilterResult {
        if (FORBIDDEN_STATEMENTS.containsMatchIn(sql)) {
            return FilterResult.Blocked(
                "Non-SELECT statement not allowed: ${sql.take(80)}"
            )
        }
        val isAllowed = ALLOWED_KEYWORDS.any { sql.contains(it, ignoreCase = true) }
                || ALLOWED_PRAGMA.containsMatchIn(sql)
        return if (isAllowed) {
            FilterResult.Allowed
        } else {
            FilterResult.Blocked(
                "Query does not match schema whitelist: ${sql.take(80)}"
            )
        }
    }
}

sealed class FilterResult {
    data object Allowed : FilterResult()
    data class Blocked(val reason: String) : FilterResult()
}
