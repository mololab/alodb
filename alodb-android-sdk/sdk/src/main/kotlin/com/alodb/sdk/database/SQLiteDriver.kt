package com.alodb.sdk.database

import android.content.ContentValues
import android.database.Cursor
import android.database.sqlite.SQLiteDatabase

class SQLiteDriver(private val db: SQLiteDatabase) : DatabaseDriver {

    override fun execute(sql: String): List<Map<String, Any?>> {
        val cursor: Cursor = db.rawQuery(sql, null)
        return cursor.use { c ->
            val results = mutableListOf<Map<String, Any?>>()
            val columnNames = c.columnNames
            while (c.moveToNext()) {
                val row = mutableMapOf<String, Any?>()
                for (i in columnNames.indices) {
                    row[columnNames[i]] = when (c.getType(i)) {
                        Cursor.FIELD_TYPE_NULL -> null
                        Cursor.FIELD_TYPE_INTEGER -> c.getLong(i)
                        Cursor.FIELD_TYPE_FLOAT -> c.getDouble(i)
                        Cursor.FIELD_TYPE_STRING -> c.getString(i)
                        Cursor.FIELD_TYPE_BLOB -> c.getBlob(i)
                        else -> c.getString(i)
                    }
                }
                results.add(row)
            }
            results
        }
    }

    override fun getDatabaseName(): String = db.path?.substringAfterLast('/')?.removeSuffix(".db") ?: "main"

    override fun insert(table: String, rows: List<Map<String, Any?>>) {
        db.beginTransaction()
        try {
            for (row in rows) {
                val values = ContentValues(row.size)
                for ((key, value) in row) {
                    when (value) {
                        null -> values.putNull(key)
                        is String -> values.put(key, value)
                        is Int -> values.put(key, value)
                        is Long -> values.put(key, value)
                        is Float -> values.put(key, value)
                        is Double -> values.put(key, value)
                        is Boolean -> values.put(key, if (value) 1 else 0)
                        is ByteArray -> values.put(key, value)
                        else -> values.put(key, value.toString())
                    }
                }
                db.insertWithOnConflict(table, null, values, SQLiteDatabase.CONFLICT_REPLACE)
            }
            db.setTransactionSuccessful()
        } finally {
            db.endTransaction()
        }
    }

    override fun clearTable(table: String): Int {
        return db.delete(table, null, null)
    }

    override fun tableRowCount(table: String): Int {
        val cursor = db.rawQuery("SELECT COUNT(*) FROM \"$table\"", null)
        return cursor.use { c ->
            if (c.moveToFirst()) c.getInt(0) else 0
        }
    }
}
