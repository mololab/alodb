package com.alodb.sample

import android.app.Application
import androidx.room.Room

class SampleApp : Application() {

    lateinit var database: AppDatabase
        private set

    override fun onCreate() {
        super.onCreate()
        database = Room.databaseBuilder(this, AppDatabase::class.java, "sample-bank")
            .fallbackToDestructiveMigration()
            .build()
    }
}
