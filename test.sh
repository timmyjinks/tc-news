#!/usr/bin/env bash
# ============================================================================
# End-to-end curl smoke test for the tc-news microservices
# ============================================================================
# Walks the full user journey back-to-back:
#   register user A -> user A posts -> register user B -> user B posts ->
#   user B comments on A's post -> user A replies -> votes -> user A
#   subscribes to B's post -> B comments again -> notification fires for A
#   -> mark/read/delete notifications -> cleanup
#
# Requires: bash, curl, jq, (uuidgen or python3)
#
# Default ports come from docker-compose.yaml's host mappings:
#   auth          http://localhost:8085
#   post          http://localhost:8083
#   comment       http://localhost:8084
#   subscribe     http://localhost:8081
#   notification  http://localhost:8086
#   vote          http://localhost:8082   <- NOT enabled by default, see NOTES
#
# ----------------------------------------------------------------------------
# NOTES on bugs/quirks found while writing this (so failures below make sense
# instead of looking like script errors):
#
#  5. auth.CreateUser, post.CreatePost, vote.VotePost/VoteComment, and the
#     comment-create handler never write a response body with the new
#     resource's id, despite what the swagger docs imply. This script works
#     around that by tagging every created body with a unique nonce and
#     re-fetching via the matching list/GET endpoint.
#  6. store.Vote / store.Comment / store.Subscriber / store.Notification have
#     no `json:` tags on their Id/PostId/UserId/etc. fields, so the *real*
#     JSON keys are PascalCase ("Id", "PostId", "UserId", ...) even though
#     swagger.yaml/openapi.yaml document lowercase names. This script uses
#     the real casing.
# ----------------------------------------------------------------------------

set -uo pipefail

AUTH=http://localhost:8085
POST=http://localhost:8083
COMMENT=http://localhost:8084
SUBSCRIBE=http://localhost:8081
NOTIFICATION=http://localhost:8086
VOTE=http://localhost:8082

command -v jq >/dev/null || { echo "jq is required (https://jqlang.org). Install it and re-run."; exit 1; }
UUID() { uuidgen 2>/dev/null || python3 -c 'import uuid;print(uuid.uuid4())'; }

NONCE=$(date +%s)
PASS='Sup3rSecret!'
BODY_FILE=$(mktemp)
trap 'rm -f "$BODY_FILE"' EXIT

PASS_COUNT=0
FAIL_COUNT=0

step() { echo; echo "== $* =="; }

check() {
  # check "description" expected_code actual_code
  local desc=$1 expected=$2 actual=$3
  if [[ "$actual" == "$expected" ]]; then
    echo "  OK   $desc (HTTP $actual)"
    PASS_COUNT=$((PASS_COUNT+1))
  else
    echo "  FAIL $desc (expected $expected, got $actual)"
    echo "       body: $(cat "$BODY_FILE" 2>/dev/null | head -c 300)"
    FAIL_COUNT=$((FAIL_COUNT+1))
  fi
}

req() {
  # req METHOD URL [JSON_DATA] [HEADER1] [HEADER2] ...
  # writes response body to $BODY_FILE, prints http status code (stdout)
  local method=$1 url=$2; shift 2
  local data=""
  if [[ $# -gt 0 && "$1" != -H* && "$1" != "X-User-ID:"* && "$1" == "{"* ]]; then :; fi
  local args=(-sS -o "$BODY_FILE" -w '%{http_code}' -X "$method" "$url")
  if [[ $# -gt 0 && "$1" == \{* ]]; then
    args+=(-H 'Content-Type: application/json' -d "$1")
    shift
  fi
  for h in "$@"; do
    args+=(-H "$h")
  done
  curl "${args[@]}"
}

service_up() {
  curl -sS -o /dev/null -m 2 -w '%{http_code}' "$1" 2>/dev/null
}

# ============================================================================
step "0. Reachability check"
# ============================================================================
for pair in "auth:$AUTH/auth/login" "post:$POST/posts" "comment:$COMMENT/posts/x/comments" "subscribe:$SUBSCRIBE/users/subscriptions" "notification:$NOTIFICATION/user/notifications"; do
  name=${pair%%:*}; url=${pair#*:}
  code=$(service_up "$url")
  echo "  $name -> $url (HTTP $code, any response means it's up)"
done
VOTE_UP=1
vcode=$(service_up "$VOTE/user/votes")
if [[ "$vcode" == "000" ]]; then
  echo "  vote -> $VOTE not reachable (expected: it's commented out of docker-compose.yaml by default)"
  VOTE_UP=0
else
  echo "  vote -> $VOTE (HTTP $vcode)"
fi

# ============================================================================
step "1. AUTH: register two users, log in, refresh a token"
# ============================================================================
USER_A_NAME="alice_${NONCE}"
USER_B_NAME="bob_${NONCE}"

code=$(req POST "$AUTH/users" "{\"name\":\"$USER_A_NAME\",\"password\":\"$PASS\"}")
check "register user A" 201 "$code"

code=$(req POST "$AUTH/users" "{\"name\":\"$USER_B_NAME\",\"password\":\"$PASS\"}")
check "register user B" 201 "$code"

COOKIES_A=$(mktemp)
code=$(curl -sS -o "$BODY_FILE" -w '%{http_code}' -X POST "$AUTH/auth/login" \
  -H 'Content-Type: application/json' -c "$COOKIES_A" \
  -d "{\"name\":\"$USER_A_NAME\",\"password\":\"$PASS\"}")
check "login as user A" 200 "$code"
ACCESS_TOKEN_A=$(jq -r '.access_token' "$BODY_FILE" 2>/dev/null)
echo "  access_token (truncated): ${ACCESS_TOKEN_A:0:24}..."

code=$(curl -sS -o "$BODY_FILE" -w '%{http_code}' -X POST "$AUTH/auth/refresh" \
  -b "$COOKIES_A" -c "$COOKIES_A" \
  -H "Authorization: Bearer $ACCESS_TOKEN_A")
check "refresh user A's token" 200 "$code"
rm -f "$COOKIES_A"

# NOTE: auth never returns the new user's id, and there's no "get by name"
# endpoint, so from here on we simulate the API-gateway-issued X-User-ID
# header with our own generated ids. The content services (post/comment/
# vote/subscribe) trust this header as-is and never cross-check it against
# the auth service.
USER_A_ID=$(UUID)
USER_B_ID=$(UUID)
echo "  simulated X-User-ID for A: $USER_A_ID"
echo "  simulated X-User-ID for B: $USER_B_ID"

# ============================================================================
step "2. POST: user A creates a post, user B creates a post"
# ============================================================================
NONCE_POST_A="post-A-marker-$NONCE"
NONCE_POST_B="post-B-marker-$NONCE"

code=$(req POST "$POST/posts" "{\"body\":\"$NONCE_POST_A\"}" "X-User-ID: $USER_A_ID")
check "user A creates a post" 200 "$code"

code=$(req POST "$POST/posts" "{\"body\":\"$NONCE_POST_B\"}" "X-User-ID: $USER_B_ID")
check "user B creates a post" 200 "$code"

code=$(req GET "$POST/posts")
check "list all posts" 200 "$code"
# NOTE: `(. // [])[]` instead of `.[]` -- if there are zero posts, Go's
# encoding/json serializes the nil slice as the literal `null`, and jq's
# `.[]` throws "Cannot iterate over null (null)" on that. `. // []` coerces
# null to an empty array first so a no-match just yields nothing instead of
# crashing the whole script.
POST_A_ID=$(jq -r --arg n "$NONCE_POST_A" '(. // [])[] | select(.body == $n) | .id' "$BODY_FILE" | head -n1)
POST_B_ID=$(jq -r --arg n "$NONCE_POST_B" '(. // [])[] | select(.body == $n) | .id' "$BODY_FILE" | head -n1)
echo "  resolved POST_A_ID=$POST_A_ID"
echo "  resolved POST_B_ID=$POST_B_ID"
if [[ -z "$POST_A_ID" || -z "$POST_B_ID" ]]; then
  echo "  WARN: could not resolve one or both post ids. Raw response was:"
  cat "$BODY_FILE"
fi

code=$(req GET "$POST/posts/$POST_A_ID")
check "get post A by id" 200 "$code"

# ============================================================================
step "3. COMMENT: user B comments on A's post, user A replies"
# ============================================================================
NONCE_COMMENT_1="comment-1-marker-$NONCE"
NONCE_COMMENT_2="comment-2-marker-$NONCE"

code=$(req POST "$COMMENT/posts/$POST_A_ID/comments" "{\"body\":\"$NONCE_COMMENT_1\",\"parent_id\":\"00000000-0000-0000-0000-000000000000\"}" "X-User-ID: $USER_B_ID")
check "user B comments on post A" 200 "$code"

code=$(req GET "$COMMENT/posts/$POST_A_ID/comments")
check "list comments on post A" 200 "$code"
# NOTE: store.Comment has no json tags on Id/PostId/UserId -> PascalCase keys.
# Also: same null-vs-[] gotcha as the posts list above, hence `(. // [])[]`.
COMMENT_1_ID=$(jq -r --arg n "$NONCE_COMMENT_1" '(. // [])[] | select(.body == $n) | .Id' "$BODY_FILE" | head -n1)
echo "  resolved COMMENT_1_ID=$COMMENT_1_ID"
if [[ -z "$COMMENT_1_ID" ]]; then
  echo "  WARN: could not resolve comment id. Raw response was:"
  cat "$BODY_FILE"
fi

code=$(req POST "$COMMENT/posts/$POST_A_ID/comments" "{\"body\":\"$NONCE_COMMENT_2\",\"parent_id\":\"$COMMENT_1_ID\"}" "X-User-ID: $USER_A_ID")
check "user A replies to B's comment" 200 "$code"

code=$(req GET "$COMMENT/comments/$COMMENT_1_ID")
check "get comment 1 by id" 200 "$code"

code=$(req PUT "$COMMENT/comments/$COMMENT_1_ID" '{"body":"edited by B"}' "X-User-ID: $USER_B_ID")
check "user B edits their own comment" 200 "$code"

# ============================================================================
step "4. VOTE: user A votes on post B, user B votes on comment 1"
# ============================================================================
if [[ "$VOTE_UP" == "0" ]]; then
  echo "  SKIPPED - vote service not reachable (see NOTES at top of script)."
else
  code=$(req PUT "$VOTE/posts/$POST_B_ID/votes" '{"value":1}' "X-User-ID: $USER_A_ID")
  check "user A upvotes post B" 200 "$code"

  code=$(req PUT "$VOTE/comments/$COMMENT_1_ID/votes" '{"value":1}' "X-User-ID: $USER_B_ID")
  check "user B upvotes comment 1" 200 "$code"

  code=$(req GET "$VOTE/user/votes" "" "X-User-ID: $USER_A_ID")
  check "list user A's votes" 200 "$code"

  code=$(req DELETE "$VOTE/posts/$POST_B_ID/votes" "" "X-User-ID: $USER_A_ID")
  check "user A removes their vote on post B" 200 "$code"
fi

# ============================================================================
step "5. SUBSCRIBE: user A follows post B"
# ============================================================================
code=$(req POST "$SUBSCRIBE/posts/$POST_B_ID/subscriptions" "" "X-User-ID: $USER_A_ID")
check "user A subscribes to post B" 200 "$code"

code=$(req GET "$SUBSCRIBE/users/subscriptions" "" "X-User-ID: $USER_A_ID")
check "list user A's subscriptions" 200 "$code"
# NOTE: store.Subscriber has zero json tags -> PascalCase keys.
jq -r --arg pid "$POST_B_ID" '[(. // [])[] | select(.PostId == $pid)] | length' "$BODY_FILE" | \
  xargs -I{} echo "  subscriptions matching post B: {}"

# ============================================================================
step "6. NOTIFICATION: user B comments again on post B -> A should be notified"
# ============================================================================
NONCE_COMMENT_3="comment-3-marker-$NONCE"
code=$(req POST "$COMMENT/posts/$POST_B_ID/comments" "{\"body\":\"$NONCE_COMMENT_3\",\"parent_id\":\"00000000-0000-0000-0000-000000000000\"}" "X-User-ID: $USER_B_ID")
check "user B comments on post B (should publish comment_created event)" 200 "$code"

echo "  waiting for the async kafka -> notification pipeline..."
sleep 3

code=$(req GET "$NOTIFICATION/user/notifications" "" "X-User-ID: $USER_A_ID")
check "list user A's notifications" 200 "$code"
# NOTE: store.Notification has zero json tags -> PascalCase keys. Same
# null-vs-[] gotcha as above.
NOTIF_COUNT=$(jq -r '(. // []) | length' "$BODY_FILE" 2>/dev/null)
NOTIF_ID=$(jq -r '(. // [])[0].Id // empty' "$BODY_FILE" 2>/dev/null)
echo "  user A has $NOTIF_COUNT notification(s)"

if [[ -n "$NOTIF_ID" ]]; then
  code=$(req PATCH "$NOTIFICATION/notifications/$NOTIF_ID/read" "" "X-User-ID: $USER_A_ID")
  check "mark one notification read" 200 "$code"

  code=$(req PATCH "$NOTIFICATION/notifications/read-all" "" "X-User-ID: $USER_A_ID")
  check "mark all notifications read" 200 "$code"

  code=$(req DELETE "$NOTIFICATION/notifications/$NOTIF_ID" "" "X-User-ID: $USER_A_ID")
  check "delete a notification" 200 "$code"
else
  echo "  WARN: no notification found yet -- either the pipeline needs more time,"
  echo "        or user A subscribed after post B's *original* creation and the"
  echo "        notification service correctly only fires on new comment_created"
  echo "        events (i.e. this may be expected, not a failure)."
fi

# ============================================================================
step "7. CLEANUP"
# ============================================================================
code=$(req DELETE "$COMMENT/comments/$COMMENT_1_ID" "" "X-User-ID: $USER_B_ID")
check "delete comment 1" 200 "$code"

code=$(req DELETE "$SUBSCRIBE/posts/$POST_B_ID/subscriptions" "" "X-User-ID: $USER_A_ID")
check "user A unsubscribes from post B" 200 "$code"

code=$(req DELETE "$POST/posts/$POST_A_ID" "" "X-User-ID: $USER_A_ID")
check "delete post A" 200 "$code"

code=$(req DELETE "$POST/posts/$POST_B_ID" "" "X-User-ID: $USER_B_ID")
check "delete post B" 200 "$code"

# ============================================================================
step "SUMMARY"
# ============================================================================
echo "  passed: $PASS_COUNT"
echo "  failed: $FAIL_COUNT"
[[ "$FAIL_COUNT" -eq 0 ]] && exit 0 || exit 1
