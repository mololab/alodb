package com.alodb.sample

import android.os.Bundle
import android.view.inputmethod.EditorInfo
import android.widget.EditText
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.alodb.sdk.AloDBClient
import com.alodb.sdk.AloDBListener
import com.alodb.sdk.database.RoomDriver
import com.alodb.sdk.model.GeneratedSQL
import com.alodb.sdk.session.ClearResult
import com.google.android.material.button.MaterialButton
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class MainActivity : AppCompatActivity() {

    private lateinit var client: AloDBClient
    private lateinit var adapter: ChatAdapter
    private lateinit var recycler: RecyclerView
    private lateinit var editMessage: EditText

    private val messages = mutableListOf<ChatMessage>()
    private val streamingBuffer = StringBuilder()
    private var streamingMessageIndex = -1
    private var thinkingMessageIndex = -1

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        recycler = findViewById(R.id.recyclerChat)
        editMessage = findViewById(R.id.editMessage)
        val btnSend = findViewById<MaterialButton>(R.id.btnSend)

        adapter = ChatAdapter(messages)
        recycler.layoutManager = LinearLayoutManager(this).apply {
            stackFromEnd = true
        }
        recycler.adapter = adapter

        val app = application as SampleApp
        val sqliteDb = app.database.openHelper.writableDatabase
        val driver = RoomDriver(sqliteDb)

        client = AloDBClient.Builder()
            .serverUrl("ws://your-server:8080/v1/agent/stream")
            .apiKey("your-gemini-api-key")
            .database(driver)
            .addTrigger("credits_page", "Kullanıcı kredi sayfasını açtı. Kredi özetini göster.")
            .addTrigger("home_page", "Kullanıcı ana sayfaya döndü. Bugünkü işlemleri göster.")
            .addTrigger("campaign_page", "Aktif kampanyaları listele.")
            .listener(object : AloDBListener {
                override fun onConnected(sessionId: String) {
                    addBotMessage("Merhaba! Size nasıl yardımcı olabilirim?")
                }

                override fun onThinking(status: String) {
                    showThinking()
                }

                override fun onTextDelta(delta: String) {
                    removeThinking()
                    appendStreamingDelta(delta)
                }

                override fun onResponseComplete(queries: List<GeneratedSQL>) {
                    finalizeStreaming()
                    removeThinking()
                    autoExecuteQueries(queries)
                }

                override fun onSessionDataCleared(results: Map<String, ClearResult>) {}
                override fun onClearFailed(table: String, remainingRows: Int) {}

                override fun onSecurityViolation(blockedSql: String, reason: String) {
                    addErrorMessage("Güvenlik: Sorgu engellendi.")
                }

                override fun onError(code: String, message: String) {
                    removeThinking()
                    addErrorMessage("Hata: $message")
                }

                override fun onDisconnected() {
                    addBotMessage("Bağlantı kesildi.")
                }
            })
            .build()

        writeSampleData()
        client.connect()

        btnSend.setOnClickListener { sendMessage() }
        editMessage.setOnEditorActionListener { _, actionId, _ ->
            if (actionId == EditorInfo.IME_ACTION_SEND) {
                sendMessage()
                true
            } else false
        }
    }

    private fun sendMessage() {
        val text = editMessage.text.toString().trim()
        if (text.isEmpty()) return
        editMessage.text.clear()

        addMessage(ChatMessage(ChatMessage.Type.USER, text))
        client.chat(text)
    }

    private fun autoExecuteQueries(queries: List<GeneratedSQL>) {
        if (queries.isEmpty()) return

        lifecycleScope.launch(Dispatchers.IO) {
            for (query in queries) {
                try {
                    val rows = client.database.execute(query.query)
                    val title = "${query.title} (${rows.size} satır)"
                    withContext(Dispatchers.Main) {
                        addMessage(
                            ChatMessage(
                                type = ChatMessage.Type.BOT_QUERY_RESULT,
                                text = query.description,
                                queryResults = rows,
                                queryTitle = title,
                            )
                        )
                    }
                } catch (e: Exception) {
                    withContext(Dispatchers.Main) {
                        addErrorMessage("Sorgu çalıştırılamadı: ${e.message}")
                    }
                }
            }
        }
    }

    // -- Streaming helpers --

    private fun appendStreamingDelta(delta: String) {
        runOnUiThread {
            streamingBuffer.append(delta)
            if (streamingMessageIndex == -1) {
                streamingMessageIndex = messages.size
                messages.add(ChatMessage(ChatMessage.Type.BOT_TEXT, streamingBuffer.toString()))
                adapter.notifyItemInserted(streamingMessageIndex)
            } else {
                messages[streamingMessageIndex] = messages[streamingMessageIndex].copy(text = streamingBuffer.toString())
                adapter.notifyItemChanged(streamingMessageIndex)
            }
            scrollToBottom()
        }
    }

    private fun finalizeStreaming() {
        runOnUiThread {
            streamingBuffer.clear()
            streamingMessageIndex = -1
        }
    }

    // -- Thinking indicator --

    private fun showThinking() {
        runOnUiThread {
            if (thinkingMessageIndex != -1) return@runOnUiThread
            thinkingMessageIndex = messages.size
            messages.add(ChatMessage(ChatMessage.Type.BOT_THINKING, "Düşünüyor…"))
            adapter.notifyItemInserted(thinkingMessageIndex)
            scrollToBottom()
        }
    }

    private fun removeThinking() {
        runOnUiThread {
            if (thinkingMessageIndex == -1) return@runOnUiThread
            messages.removeAt(thinkingMessageIndex)
            adapter.notifyItemRemoved(thinkingMessageIndex)
            if (streamingMessageIndex > thinkingMessageIndex) {
                streamingMessageIndex--
            }
            thinkingMessageIndex = -1
        }
    }

    // -- Message helpers --

    private fun addBotMessage(text: String) {
        addMessage(ChatMessage(ChatMessage.Type.BOT_TEXT, text))
    }

    private fun addErrorMessage(text: String) {
        addMessage(ChatMessage(ChatMessage.Type.BOT_ERROR, text))
    }

    private fun addMessage(msg: ChatMessage) {
        runOnUiThread {
            messages.add(msg)
            adapter.notifyItemInserted(messages.size - 1)
            scrollToBottom()
        }
    }

    private fun scrollToBottom() {
        recycler.scrollToPosition(messages.size - 1)
    }

    // -- Sample data --

    private fun writeSampleData() {
        client.write(
            "transactions",
            listOf(
                mapOf("id" to 1L, "amount" to 150.0, "merchant" to "Migros", "category" to "Market", "created_at" to "2026-02-15"),
                mapOf("id" to 2L, "amount" to 45.5, "merchant" to "Shell", "category" to "Akaryakıt", "created_at" to "2026-02-16"),
                mapOf("id" to 3L, "amount" to 890.0, "merchant" to "Trendyol", "category" to "Alışveriş", "created_at" to "2026-02-20"),
                mapOf("id" to 4L, "amount" to 220.0, "merchant" to "Carrefour", "category" to "Market", "created_at" to "2026-02-21"),
                mapOf("id" to 5L, "amount" to 3500.0, "merchant" to "Apple", "category" to "Teknoloji", "created_at" to "2026-02-22"),
            ),
            sessionScoped = true,
        )
        client.write(
            "campaigns",
            listOf(
                mapOf("id" to 1L, "title" to "Market %10 İndirim", "category" to "Market", "discount" to 10, "expires_at" to "2026-03-31"),
                mapOf("id" to 2L, "title" to "Akaryakıt 50 TL Puan", "category" to "Akaryakıt", "discount" to 50, "expires_at" to "2026-04-15"),
                mapOf("id" to 3L, "title" to "Online Alışveriş %5", "category" to "Alışveriş", "discount" to 5, "expires_at" to "2026-05-01"),
            ),
            sessionScoped = true,
        )
    }

    override fun onDestroy() {
        super.onDestroy()
        client.disconnect()
    }
}
