package com.alodb.sample

import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.FrameLayout
import android.widget.TextView
import androidx.core.content.ContextCompat
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

        val ctx = holder.itemView.context
        if (isUser) {
            lp.gravity = Gravity.END
            holder.cardBubble.setCardBackgroundColor(ContextCompat.getColor(ctx, R.color.bubble_user))
            holder.tvMessage.setTextColor(ContextCompat.getColor(ctx, R.color.text_user))
        } else {
            lp.gravity = Gravity.START
            holder.cardBubble.setCardBackgroundColor(ContextCompat.getColor(ctx, R.color.bubble_bot))
            holder.tvMessage.setTextColor(ContextCompat.getColor(ctx, R.color.text_bot))
        }
        holder.cardBubble.layoutParams = lp

        when (msg.type) {
            ChatMessage.Type.BOT_QUERY_RESULT -> {
                holder.tvQueryTitle.visibility = View.VISIBLE
                holder.tvQueryTitle.text = msg.queryTitle ?: "Sorgu Sonucu"
                holder.tvQueryTitle.setTextColor(ContextCompat.getColor(ctx, R.color.text_query_title))
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
                holder.tvMessage.setTextColor(ContextCompat.getColor(ctx, R.color.text_error))
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
