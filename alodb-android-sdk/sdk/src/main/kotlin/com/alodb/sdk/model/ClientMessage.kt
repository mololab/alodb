package com.alodb.sdk.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement

@Serializable
data class ClientMessage(
    val type: String,
    val payload: JsonElement,
)

@Serializable
data class ChatPayload(
    val message: String,
    val model: String? = null,
)

@Serializable
data class QueryResultPayload(
    @SerialName("request_id") val requestId: String,
    val success: Boolean,
    val rows: JsonArray? = null,
    val error: String? = null,
)
