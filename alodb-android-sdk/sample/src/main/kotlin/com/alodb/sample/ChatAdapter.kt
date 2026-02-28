package com.alodb.sample

import android.graphics.Color
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.FrameLayout
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.google.android.material.card.MaterialCardView

class ChatAdapter(
    private val messages: MutableList<ChatMessage>,
) : RecyclerView.Adapter<ChatAdapter.ViewHolder>() {

    class ViewHolder(view: View) : RecyclerView.ViewHolder(view) {
        val cardBubble: MaterialCardView = view.findViewById(R.id.cardBubble)
        val tvMessage: TextView = view.findViewById(R.id.tvMessage)
        val tvQueryTitle: TextView = view.findViewById(R.id.tvQueryTitle)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_chat_message, parent, false)
        return ViewHolder(view)
    }

    override fun onBindViewHolder(holder: ViewHolder, position: Int) {
        val msg = messages[position]
        val lp = holder.cardBubble.layoutParams as FrameLayout.LayoutParams
        val isUser = msg.type == ChatMessage.Type.USER

        if (isUser) {
            lp.gravity = Gravity.END
            holder.cardBubble.setCardBackgroundColor(Color.parseColor("#1976D2"))
            holder.tvMessage.setTextColor(Color.WHITE)
        } else {
            lp.gravity = Gravity.START
            holder.cardBubble.setCardBackgroundColor(Color.parseColor("#F5F5F5"))
            holder.tvMessage.setTextColor(Color.parseColor("#212121"))
        }
        holder.cardBubble.layoutParams = lp

        when (msg.type) {
            ChatMessage.Type.BOT_QUERY_RESULT -> {
                holder.tvQueryTitle.visibility = View.VISIBLE
                holder.tvQueryTitle.text = msg.queryTitle ?: "Sorgu Sonucu"
                holder.tvQueryTitle.setTextColor(Color.parseColor("#1976D2"))
                holder.tvMessage.text = formatQueryResults(msg.queryResults)
                holder.tvMessage.textSize = 13f
            }

            ChatMessage.Type.BOT_THINKING -> {
                holder.tvQueryTitle.visibility = View.GONE
                holder.tvMessage.text = msg.text
                holder.tvMessage.textSize = 15f
            }

            ChatMessage.Type.BOT_ERROR -> {
                holder.tvQueryTitle.visibility = View.GONE
                holder.tvMessage.text = msg.text
                holder.tvMessage.setTextColor(Color.parseColor("#D32F2F"))
                holder.tvMessage.textSize = 14f
            }

            else -> {
                holder.tvQueryTitle.visibility = View.GONE
                holder.tvMessage.text = msg.text
                holder.tvMessage.textSize = 15f
            }
        }
    }

    override fun getItemCount(): Int = messages.size

    private fun formatQueryResults(rows: List<Map<String, Any?>>?): String {
        if (rows.isNullOrEmpty()) return "Sonuç bulunamadı."

        val sb = StringBuilder()
        val columns = rows.first().keys.toList()

        for ((i, row) in rows.withIndex()) {
            if (i > 0) sb.append("\n─────────────\n")
            for (col in columns) {
                val value = row[col]?.toString() ?: "—"
                sb.append("$col: $value\n")
            }
        }
        return sb.toString().trimEnd()
    }
}
