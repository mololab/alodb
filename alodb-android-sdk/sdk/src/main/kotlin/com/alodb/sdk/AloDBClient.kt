package com.alodb.sdk

import com.alodb.sdk.database.DatabaseDriver
import com.alodb.sdk.session.SessionDataManager
import com.alodb.sdk.websocket.WebSocketManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class AloDBClient private constructor(
    private val config: AloDBConfig,
    private val driver: DatabaseDriver,
    private val listener: AloDBListener,
    private val triggers: Map<String, String>,
) {
    val database: DatabaseDriver get() = driver

    private val sessionManager = SessionDataManager(driver)
    private val clientScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var wsManager: WebSocketManager? = null

    fun connect() {
        wsManager?.disconnect()
        wsManager = WebSocketManager(
            serverUrl = config.serverUrl,
            apiKey = config.apiKey,
            model = config.model,
            driver = driver,
            listener = listener,
        )
        wsManager?.connect()
    }

    fun chat(message: String) {
        wsManager?.sendChat(message)
    }

    fun trigger(name: String) {
        val message = triggers[name]
            ?: throw IllegalArgumentException("Unknown trigger: $name. Registered: ${triggers.keys}")
        wsManager?.sendChat(message)
    }

    fun write(table: String, rows: List<Map<String, Any?>>, sessionScoped: Boolean = false) {
        driver.insert(table, rows)
        if (sessionScoped) {
            sessionManager.register(table)
        }
    }

    fun clearSessionData(): Map<String, com.alodb.sdk.session.ClearResult> {
        val results = sessionManager.clearAll()
        listener.onSessionDataCleared(results)
        for ((table, result) in results) {
            if (!result.success) {
                listener.onClearFailed(table, result.remainingRows)
            }
        }
        return results
    }

    fun disconnect() {
        wsManager?.disconnect()
        wsManager = null
        clientScope.launch { clearSessionData() }
    }

    fun logout() {
        wsManager?.disconnect()
        wsManager = null
        clientScope.launch { clearSessionData() }
    }

    class Builder {
        private var serverUrl: String? = null
        private var apiKey: String? = null
        private var model: String? = null
        private var driver: DatabaseDriver? = null
        private var listener: AloDBListener? = null
        private val triggers = mutableMapOf<String, String>()

        fun serverUrl(url: String) = apply { this.serverUrl = url }
        fun apiKey(key: String) = apply { this.apiKey = key }
        fun model(model: String) = apply { this.model = model }
        fun database(driver: DatabaseDriver) = apply { this.driver = driver }
        fun listener(listener: AloDBListener) = apply { this.listener = listener }

        fun addTrigger(name: String, message: String) = apply {
            triggers[name] = message
        }

        fun build(): AloDBClient {
            return AloDBClient(
                config = AloDBConfig(
                    serverUrl = requireNotNull(serverUrl) { "serverUrl is required" },
                    apiKey = requireNotNull(apiKey) { "apiKey is required" },
                    model = model,
                ),
                driver = requireNotNull(driver) { "database driver is required" },
                listener = requireNotNull(listener) { "listener is required" },
                triggers = triggers.toMap(),
            )
        }
    }
}
