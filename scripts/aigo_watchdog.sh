#!/bin/bash
# AIGO Watchdog - monitors for new messages and broadcasts
# Runs as a cron job (no_agent=true), outputs ONLY when there's something new
# Database direct approach — no JWT tokens needed

# === Config ===
STATE_FILE="/tmp/aigo_watchdog_state.txt"
BUYER_ID="00de712d-06a9-40d5-8d99-8fbde7b6df45"
DB="deploy-postgres-1"

# Get DB user from docker-compose or use default
DB_USER="aigo"
DB_NAME="aigo"

# Load last seen max message ID
load_state() {
    if [ -f "$STATE_FILE" ]; then
        . "$STATE_FILE" 2>/dev/null
    fi
    LAST_MSG_ID=${LAST_MSG_ID:-0}
    LAST_BC_ID=${LAST_BC_ID:-0}
}

save_state() {
    echo "LAST_MSG_ID=$1" > "$STATE_FILE"
    echo "LAST_BC_ID=$2" >> "$STATE_FILE"
}

# === SQL helpers ===
psql_cmd() {
    docker exec "$DB" psql -U "$DB_USER" -d "$DB_NAME" -t -A "$@"
}

# === Main ===
load_state

OUTPUT=""

# --- 1. Check new messages for buyer (inbox) ---
SQL="SELECT COUNT(*) FROM messages WHERE receiver_id='$BUYER_ID'"
[ "$LAST_MSG_ID" != "0" ] && SQL="$SQL AND id::text > '$LAST_MSG_ID'"
MSG_COUNT=$(psql_cmd -c "$SQL" 2>/dev/null | tr -d ' ')

if [ -n "$MSG_COUNT" ] && [ "$MSG_COUNT" -gt 0 ]; then
    # Fetch new messages
    SQL2="SELECT id, subject, body, sender_id, to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') 
          FROM messages WHERE receiver_id='$BUYER_ID'"
    [ "$LAST_MSG_ID" != "0" ] && SQL2="$SQL2 AND id::text > '$LAST_MSG_ID'"
    SQL2="$SQL2 ORDER BY created_at ASC"
    
    OUTPUT="${OUTPUT}
━━━ 📬 新私信 ━━━"
    
    # Process each message
    while IFS='|' read -r msg_id subject body sender_id created_at; do
        [ -z "$msg_id" ] && continue
        SENDER_SHORT=$(echo "$sender_id" | head -c 12)
        SUBJ="${subject:-(无主题)}"
        BODY_TRUNC=$(echo "$body" | head -c 300)
        OUTPUT="${OUTPUT}

• 来自: \`${SENDER_SHORT}...\`
  主题: ${SUBJ}
  时间: ${created_at}
  内容: ${BODY_TRUNC}"
        LAST_MSG_ID="$msg_id"
    done < <(psql_cmd -F'|' -c "$SQL2" 2>/dev/null)
fi

# --- 2. Check new broadcasts ---
SQL_BC="SELECT id, title, content, type, to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') 
         FROM broadcasts WHERE status='active'"
[ "$LAST_BC_ID" != "0" ] && SQL_BC="$SQL_BC AND id::text > '$LAST_BC_ID'"
SQL_BC="$SQL_BC ORDER BY created_at ASC"

BC_COUNT=0
BC_OUTPUT=""
while IFS='|' read -r bc_id title content bc_type created_at; do
    [ -z "$bc_id" ] && continue
    BC_COUNT=$((BC_COUNT + 1))
    CONTENT_TRUNC=$(echo "$content" | head -c 300)
    BC_OUTPUT="${BC_OUTPUT}

• [${bc_type}] ${title}
  时间: ${created_at}
  ${CONTENT_TRUNC}"
    LAST_BC_ID="$bc_id"
done < <(psql_cmd -F'|' -c "$SQL_BC" 2>/dev/null)

if [ "$BC_COUNT" -gt 0 ]; then
    OUTPUT="${OUTPUT}
━━━ 📢 新广播 ($BC_COUNT) ━━━${BC_OUTPUT}"
fi

# Save state (use the greater of whatever we found)
save_state "$LAST_MSG_ID" "$LAST_BC_ID"

# Output only if something changed
if [ -n "$OUTPUT" ]; then
    echo "🚨 **AIGO 通知** 🚨"
    echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "$OUTPUT"
fi
