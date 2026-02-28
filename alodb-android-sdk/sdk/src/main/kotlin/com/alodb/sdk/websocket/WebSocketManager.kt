package com.alodb.sdk.websocket

import com.alodb.sdk.AloDBListener
import com.alodb.sdk.database.DatabaseDriver
import com.alodb.sdk.model.*
import com.alodb.sdk.security.FilterResult
import com.alodb.sdk.security.QuerySafetyFilter
import com.alodb.sdk.translator.SchemaQueryTranslator
import kotlinx.coroutines.*
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.*
import okhttp3.*

internal class WebSocketManager(
    private val serverUrl: String,
    private val apiKey: String,
    private val model: String?,
    private val driver: DatabaseDriver,
    private val listener: AloDBListener,
) {
    private val json = Json { ignoreUnknownKeys = true; isLenient = true }
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var webSocket: WebSocket? = null
    private val client = OkHttpClient.Builder()
        .pingInterval(30, java.util.concurrent.TimeUnit.SECONDS)
        .build()

    fun connect() {
        val urlBuilder = StringBuilder(serverUrl)
        val separator = if (serverUrl.contains("?")) "&" else "?"
        urlBuilder.append(separator).append("api_key=").append(apiKey)
        if (!model.isNullOrBlank()) {
            urlBuilder.append("&model=").append(model)
        }

        val request = Request.Builder()
            .url(urlBuilder.toString())
            .build()

        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(ws: WebSocket, response: Response) {}

            override fun onMessage(ws: WebSocket, text: String) {
                scope.launch { handleMessage(text) }
            }

            override fun onFailure(ws: WebSocket, t: Throwable, response: Response?) {
                listener.onError("websocket_error", t.message ?: "Connection failed")
                listener.onDisconnected()
            }

            override fun onClosed(ws: WebSocket, code: Int, reason: String) {
                listener.onDisconnected()
            }
        })
    }

    fun sendChat(message: String, modelOverride: String? = null) {
        val payload = buildJsonObject {
            put("message", message)
            modelOverride?.let { put("model", it) }
        }
        val msg = buildJsonObject {
            put("type", "chat")
            put("payload", payload)
        }
        webSocket?.send(json.encodeToString(msg))
    }

    fun disconnect() {
        scope.cancel()
        webSocket?.close(1000, "Client disconnect")
        webSocket = null
    }

    private suspend fun handleMessage(text: String) {
        try {
            val event = json.decodeFromString<ServerEvent>(text)
            val payloadElement = event.payload ?: return

            when (event.type) {
                "session_created" -> {
                    val p = json.decodeFromJsonElement<SessionCreatedPayload>(payloadElement)
                    listener.onConnected(p.sessionId)
                }

                "thinking" -> {
                    val p = json.decodeFromJsonElement<ThinkingPayload>(payloadElement)
                    listener.onThinking(p.status)
                }

                "query_request" -> {
                    val p = json.decodeFromJsonElement<QueryRequestPayload>(payloadElement)
                    handleQueryRequest(p)
                }

                "text_delta" -> {
                    val p = json.decodeFromJsonElement<TextDeltaPayload>(payloadElement)
                    listener.onTextDelta(p.delta)
                }

                "response_complete" -> {
                    val p = json.decodeFromJsonElement<ResponseCompletePayload>(payloadElement)
                    listener.onResponseComplete(p.queries ?: emptyList())
                }

                "error" -> {
                    val p = json.decodeFromJsonElement<ErrorPayload>(payloadElement)
                    listener.onError(p.code, p.message)
                }
            }
        } catch (e: Exception) {
            listener.onError("parse_error", "Failed to parse server message: ${e.message}")
        }
    }

    private fun handleQueryRequest(payload: QueryRequestPayload) {
        when (val filterResult = QuerySafetyFilter.check(payload.query)) {
            is FilterResult.Allowed -> executeAndSend(payload)
            is FilterResult.Blocked -> {
                sendQueryResult(payload.requestId, success = false, error = filterResult.reason)
                listener.onSecurityViolation(payload.query, filterResult.reason)
            }
        }
    }

    private fun executeAndSend(payload: QueryRequestPayload) {
        try {
            val translated = SchemaQueryTranslator.translate(payload.query, driver)
            val rawRows = driver.execute(translated.sql)
            val mappedRows = translated.postProcess?.invoke(rawRows) ?: rawRows
            sendQueryResult(payload.requestId, success = true, rows = mappedRows)
        } catch (e: Exception) {
            sendQueryResult(payload.requestId, success = false, error = e.message ?: "Query execution failed")
        }
    }

    private fun sendQueryResult(
        requestId: String,
        success: Boolean,
        rows: List<Map<String, Any?>>? = null,
        error: String? = null,
    ) {
        val payload = buildJsonObject {
            put("request_id", requestId)
            put("success", success)
            if (rows != null) {
                put("rows", JsonArray(rows.map { row ->
                    buildJsonObject {
                        for ((key, value) in row) {
                            when (value) {
                                null -> put(key, JsonNull)
                                is Number -> put(key, JsonPrimitive(value))
                                is Boolean -> put(key, JsonPrimitive(value))
                                else -> put(key, JsonPrimitive(value.toString()))
                            }
                        }
                    }
                }))
            }
            if (error != null) {
                put("error", error)
            }
        }
        val msg = buildJsonObject {
            put("type", "query_result")
            put("payload", payload)
        }
        webSocket?.send(json.encodeToString(msg))
    }
}
