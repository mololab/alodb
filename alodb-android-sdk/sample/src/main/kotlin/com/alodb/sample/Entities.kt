package com.alodb.sample

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "transactions")
data class TransactionEntity(
    @PrimaryKey val id: Long,
    val amount: Double,
    val merchant: String,
    val category: String,
    @ColumnInfo(name = "created_at") val createdAt: String,
)

@Entity(tableName = "campaigns")
data class CampaignEntity(
    @PrimaryKey val id: Long,
    val title: String,
    val category: String,
    val discount: Int,
    @ColumnInfo(name = "expires_at") val expiresAt: String,
)
