package com.alodb.sample

import androidx.room.Database
import androidx.room.RoomDatabase

@Database(
    entities = [TransactionEntity::class, CampaignEntity::class],
    version = 1,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase()
