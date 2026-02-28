package com.alodb.sample

data class ChatMessage(
    val type: Type,
    val text: String,
    val queryResults: List<Map<String, Any?>>? = null,
    val queryTitle: String? = null,
) {
    enum class Type {
        USER,
        BOT_TEXT,
        BOT_THINKING,
        BOT_QUERY_RESULT,
        BOT_ERROR,
    }
}
