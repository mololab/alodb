package com.alodb.sdk.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

@Serializable
data class ServerEvent(
    val type: String,
    @SerialName("session_id") val sessionId: String = "",
    val timestamp: String = "",
    val payload: JsonElement? = null,
)

@Serializable
data class SessionCreatedPayload(
    @SerialName("session_id") val sessionId: String,
)

@Serializable
data class ThinkingPayload(
    val status: String,
)

@Serializable
data class QueryRequestPayload(
    @SerialName("request_id") val requestId: String,
    val name: String,
    val description: String,
    val query: String,
    val step: Int,
    @SerialName("total_steps") val totalSteps: Int,
)

@Serializable
data class TextDeltaPayload(
    val delta: String,
)

@Serializable
data class ResponseCompletePayload(
    val success: Boolean,
    val message: String? = null,
    val queries: List<GeneratedSQL>? = null,
)

@Serializable
data class ErrorPayload(
    val code: String,
    val message: String,
)
