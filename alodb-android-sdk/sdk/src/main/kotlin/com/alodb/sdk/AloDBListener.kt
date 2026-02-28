package com.alodb.sdk

import com.alodb.sdk.model.GeneratedSQL
import com.alodb.sdk.session.ClearResult

interface AloDBListener {
    fun onConnected(sessionId: String)
    fun onThinking(status: String)
    fun onTextDelta(delta: String)
    fun onResponseComplete(queries: List<GeneratedSQL>)
    fun onSessionDataCleared(results: Map<String, ClearResult>)
    fun onClearFailed(table: String, remainingRows: Int)
    fun onSecurityViolation(blockedSql: String, reason: String)
    fun onError(code: String, message: String)
    fun onDisconnected()
}
