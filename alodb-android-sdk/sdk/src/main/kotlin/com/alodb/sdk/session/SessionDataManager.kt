package com.alodb.sdk.session

import com.alodb.sdk.database.DatabaseDriver

class SessionDataManager(private val driver: DatabaseDriver) {

    private val sessionTables = mutableSetOf<String>()

    fun register(table: String) {
        sessionTables.add(table)
    }

    fun registeredTables(): Set<String> = sessionTables.toSet()

    fun clearAll(): Map<String, ClearResult> {
        return sessionTables.associateWith { table ->
            try {
                val deleted = driver.clearTable(table)
                val remaining = driver.tableRowCount(table)
                ClearResult(
                    deletedRows = deleted,
                    remainingRows = remaining,
                    success = remaining == 0,
                )
            } catch (e: Exception) {
                ClearResult(
                    deletedRows = 0,
                    remainingRows = -1,
                    success = false,
                    error = e.message,
                )
            }
        }
    }
}

data class ClearResult(
    val deletedRows: Int,
    val remainingRows: Int,
    val success: Boolean,
    val error: String? = null,
)
