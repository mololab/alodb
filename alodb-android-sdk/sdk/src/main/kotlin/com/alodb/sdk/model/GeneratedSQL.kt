package com.alodb.sdk.model

import kotlinx.serialization.Serializable

@Serializable
data class GeneratedSQL(
    val title: String,
    val query: String,
    val description: String,
    val diagram: String? = null,
)
