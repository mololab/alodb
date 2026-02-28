package com.alodb.sdk

data class AloDBConfig(
    val serverUrl: String,
    val apiKey: String,
    val model: String? = null,
    val sendQueryResults: Boolean = false,
)
