package com.alodb.sdk.database

interface DatabaseDriver {
    fun execute(sql: String): List<Map<String, Any?>>
    fun getDatabaseName(): String
    fun insert(table: String, rows: List<Map<String, Any?>>)
    fun clearTable(table: String): Int
    fun tableRowCount(table: String): Int
}
