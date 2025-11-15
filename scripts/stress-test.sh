#!/usr/bin/env bash
set -euo pipefail

TOTAL_REQUESTS=${TOTAL_REQUESTS:-1000}
CONCURRENCY=${CONCURRENCY:-100}
HOST=${HOST:-"http://localhost:8080"}
SUBJECT_TEMPLATE=${SUBJECT_TEMPLATE:-"stress-user"}
BASE_SCORE=${BASE_SCORE:-100}
LEADERBOARD_FROM=${LEADERBOARD_FROM:-0}
LEADERBOARD_LIMIT=${LEADERBOARD_LIMIT:-1000}

random_quiz_id() {
  printf "quiz-%s" "$(openssl rand -hex 4)"
}

run_request() {
  local idx=$1
  local quiz_id="$(random_quiz_id)"
  local score=$((BASE_SCORE + idx))

  local token=$(go run ./cmd/gen-token -sub "stress-${idx}" | awk -F"'" 'NR==2{print $2}')

  curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${HOST}/user/quiz/${quiz_id}/join" \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    -d "{}" >/dev/null

  curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${HOST}/user/quiz/${quiz_id}/submit" \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    -d "{\"score\":${score}}" >/dev/null
}

export -f run_request random_quiz_id
export HOST BASE_SCORE

redis-cli -h localhost -p 6379 -n 0 DEL emu-game:scores >/dev/null 2>&1 || true

start_time=$(date +%s)
echo "Starting stress test at $(date -r $start_time '+%Y-%m-%d %H:%M:%S')"

seq 1 "$TOTAL_REQUESTS" | xargs -I{} -P "$CONCURRENCY" bash -c 'run_request "$@"' _ {}

end_time=$(date +%s)
echo "Completed at $(date -r $end_time '+%Y-%m-%d %H:%M:%S')"
echo "Duration: $((end_time - start_time)) seconds"

LEADERBOARD=$(curl -s -X GET "${HOST}/leaderboard" \
  -H "Content-Type: application/json" \
  -d "{\"from\":${LEADERBOARD_FROM},\"limit\":${LEADERBOARD_LIMIT}}")

LEADERBOARD_COUNT=$(echo "$LEADERBOARD" | python3 -c "import sys, json; data = json.load(sys.stdin); print(len(data) if isinstance(data, list) else 0)" 2>/dev/null || echo "0")

echo "Leaderboard entries: ${LEADERBOARD_COUNT}"
echo "Completed ${TOTAL_REQUESTS} join+submit flows with random users/quizzes (concurrency=${CONCURRENCY})"
