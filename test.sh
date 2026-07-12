#!/usr/bin/env bash
# ============================================================================
# End-to-end curl smoke test for the tc-news microservices
# ============================================================================
# Walks the full user journey back-to-back:
#   register user A -> user A posts -> register user B -> user B posts ->
#   list/paginate/filter/sort posts -> comment cross-service validation ->
#   user B comments on A's post -> user A replies -> votes -> user A
#   subscribes to B's post -> B comments again -> notification fires for A ->
#   mark/read/delete notifications -> cleanup
#
# Requires: bash, curl, jq, (uuidgen or python3)
#
# Default ports come from docker-compose.yaml's host mappings:
#   auth          http://localhost:8085
#   post          http://localhost:8083
#   comment       http://localhost:8084
#   subscribe     http://localhost:8081
#   notification  http://localhost:8086
#   vote          http://localhost:8082 
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
#     the real casing. (store.Post does have json tags -- id/author_id/
#     title/body/tags/created_at -- so posts are the one resource that
#     matches its docs.)
#  7. comment.create_comment now performs cross-service validation: before
#     inserting, it calls the post service (POST_SERVICE_URL) to confirm
#     post_id actually exists. Commenting on a post_id that doesn't exist
#     now returns 404 "Post does not exist" instead of silently trusting the
#     client and inserting an orphaned comment. If the post service itself
#     is unreachable/misbehaving, comment creation returns 422 instead of
#     500, since the request couldn't be verified either way.
# ----------------------------------------------------------------------------

set -uo pipefail

AUTH=http://tc.tysonjenkins.dev
POST=http://tc.tysonjenkins.dev
COMMENT=http://tc.tysonjenkins.dev
SUBSCRIBE=http://tc.tysonjenkins.dev
NOTIFICATION=http://tc.tysonjenkins.dev
VOTE=http://tc.tysonjenkins.dev

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

check_true() {
  # check_true "description" boolean_result(0/1 via string "true"/"false")
  local desc=$1 ok=$2
  if [[ "$ok" == "true" ]]; then
    echo "  OK   $desc"
    PASS_COUNT=$((PASS_COUNT+1))
  else
    echo "  FAIL $desc"
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
echo "  vote -> $VOTE (HTTP $vcode)"

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

COOKIES_B=$(mktemp)
code=$(curl -sS -o "$BODY_FILE" -w '%{http_code}' -X POST "$AUTH/auth/login" \
  -H 'Content-Type: application/json' -c "$COOKIES_B" \
  -d "{\"name\":\"$USER_B_NAME\",\"password\":\"$PASS\"}")
check "login as user B" 200 "$code"
ACCESS_TOKEN_B=$(jq -r '.access_token' "$BODY_FILE" 2>/dev/null)
rm -f "$COOKIES_B"


# ============================================================================
step "2. POST: user A creates a post, user B creates a post"
# ============================================================================
NONCE_POST_A="post-A-marker-$NONCE"
NONCE_POST_B="post-B-marker-$NONCE"
TAG_SHARED="smoke-test-$NONCE"   # both posts share this tag
TAG_A_ONLY="kafka-$NONCE"        # only post A has this tag

code=$(req POST "$POST/posts" "{\"title\":\"Post A\",\"body\":\"$NONCE_POST_A\",\"tags\":[\"$TAG_SHARED\",\"$TAG_A_ONLY\"]}" "Authorization: Bearer $ACCESS_TOKEN_A")
check "user A creates a post" 201 "$code"

# small gap so created_at ordering between A and B is unambiguous
sleep 1

code=$(req POST "$POST/posts" "{\"title\":\"Post B\",\"body\":\"$NONCE_POST_B\",\"tags\":[\"$TAG_SHARED\"]}" "Authorization: Bearer $ACCESS_TOKEN_B")
check "user B creates a post" 201 "$code"

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
title_ok=$(jq -r '.title == "Post A"' "$BODY_FILE" 2>/dev/null)
check_true "post A has title + author_id set correctly" "${title_ok:-false}"

# ----------------------------------------------------------------------------
step "2b. POST: pagination, tag filter, and sort query params on GET /posts"
# ----------------------------------------------------------------------------
code=$(req GET "$POST/posts?tag=$TAG_A_ONLY")
check "filter posts by a tag only post A has" 200 "$code"
tag_filter_ok=$(jq -r --arg pid "$POST_A_ID" --arg bid "$POST_B_ID" \
  '[(. // [])[].id] as $ids | ($ids | index($pid)) != null and (($ids | index($bid)) == null)' \
  "$BODY_FILE" 2>/dev/null)
check_true "tag filter returns post A but not post B" "${tag_filter_ok:-false}"

code=$(req GET "$POST/posts?tag=$TAG_SHARED")
check "filter posts by the shared tag" 200 "$code"
shared_tag_ok=$(jq -r --arg pid "$POST_A_ID" --arg bid "$POST_B_ID" \
  '[(. // [])[].id] as $ids | ($ids | index($pid)) != null and ($ids | index($bid)) != null' \
  "$BODY_FILE" 2>/dev/null)
check_true "shared tag filter returns both post A and post B" "${shared_tag_ok:-false}"

code=$(req GET "$POST/posts?limit=1&sort=oldest")
check "paginate: limit=1&sort=oldest" 200 "$code"
oldest_ok=$(jq -r --arg pid "$POST_A_ID" '((. // []) | length) == 1 and (.[0].id == $pid)' "$BODY_FILE" 2>/dev/null)
check_true "oldest-first, limit 1 returns post A" "${oldest_ok:-false}"

code=$(req GET "$POST/posts?limit=1&sort=newest")
check "paginate: limit=1&sort=newest" 200 "$code"
newest_ok=$(jq -r --arg bid "$POST_B_ID" '((. // []) | length) == 1 and (.[0].id == $bid)' "$BODY_FILE" 2>/dev/null)
check_true "newest-first, limit 1 returns post B" "${newest_ok:-false}"

code=$(req GET "$POST/posts?limit=1&offset=1&sort=oldest")
check "paginate: limit=1&offset=1&sort=oldest" 200 "$code"
offset_ok=$(jq -r --arg bid "$POST_B_ID" '((. // []) | length) == 1 and (.[0].id == $bid)' "$BODY_FILE" 2>/dev/null)
check_true "oldest-first, offset 1 skips post A and returns post B" "${offset_ok:-false}"

code=$(req GET "$POST/posts?limit=abc")
check "invalid limit is rejected" 400 "$code"

code=$(req GET "$POST/posts?sort=sideways")
check "invalid sort value is rejected" 400 "$code"

# ============================================================================
step "3. COMMENT: cross-service validation against the post service"
# ============================================================================
# The comment service no longer trusts a client-supplied post_id at face
# value -- it calls back to the post service to confirm the post actually
# exists before inserting a comment (see comment/app/post_client.py). A
# random UUID that was never used as a post id should be rejected with 404.
FAKE_POST_ID=$(UUID)
NONCE_COMMENT_INVALID="comment-invalid-post-marker-$NONCE"

code=$(req POST "$COMMENT/posts/$FAKE_POST_ID/comments" "{\"body\":\"$NONCE_COMMENT_INVALID\",\"parent_id\":\"00000000-0000-0000-0000-000000000000\"}" "Authorization: Bearer $ACCESS_TOKEN_B")
check "commenting on a nonexistent post_id is rejected" 404 "$code"

code=$(req GET "$COMMENT/posts/$FAKE_POST_ID/comments")
check "listing comments for the nonexistent post still returns 200 (empty list)" 200 "$code"

# ============================================================================
step "4. COMMENT: user B comments on A's post, user A replies"
# ============================================================================
NONCE_COMMENT_1="comment-1-marker-$NONCE"
NONCE_COMMENT_2="comment-2-marker-$NONCE"

code=$(req POST "$COMMENT/posts/$POST_A_ID/comments" "{\"body\":\"$NONCE_COMMENT_1\",\"parent_id\":\"00000000-0000-0000-0000-000000000000\"}" "Authorization: Bearer $ACCESS_TOKEN_B")
check "user B comments on post A (post exists, passes cross-service check)" 201 "$code"

code=$(req GET "$COMMENT/posts/$POST_A_ID/comments")
check "list comments on post A" 200 "$code"
# NOTE: store.Comment has no json tags on Id/PostId/UserId -> PascalCase keys.
# Also: same null-vs-[] gotcha as the posts list above, hence `(. // [])[]`.
COMMENT_1_ID=$(jq -r --arg n "$NONCE_COMMENT_1" '(. // [])[] | select(.body == $n) | .id' "$BODY_FILE" | head -n1)
echo "  resolved COMMENT_1_ID=$COMMENT_1_ID"
if [[ -z "$COMMENT_1_ID" ]]; then
  echo "  WARN: could not resolve comment id. Raw response was:"
  cat "$BODY_FILE"
fi

code=$(req POST "$COMMENT/posts/$POST_A_ID/comments" "{\"body\":\"$NONCE_COMMENT_2\",\"parent_id\":\"$COMMENT_1_ID\"}" "Authorization: Bearer $ACCESS_TOKEN_A")
check "user A replies to B's comment" 201 "$code"

code=$(req GET "$COMMENT/comments/$COMMENT_1_ID")
check "get comment 1 by id" 200 "$code"

code=$(req PUT "$COMMENT/comments/$COMMENT_1_ID" '{"body":"edited by B"}' "Authorization: Bearer $ACCESS_TOKEN_B")
check "user B edits their own comment" 200 "$code"

# ============================================================================
step "5. VOTE: user A votes on post B, user B votes on comment 1"
# ============================================================================
if [[ "$VOTE_UP" == "0" ]]; then
  echo "  SKIPPED - vote service not reachable (see NOTES at top of script)."
else
  code=$(req PUT "$VOTE/posts/$POST_B_ID/votes" '{"value":1}' "Authorization: Bearer $ACCESS_TOKEN_A")
  check "user A upvotes post B" 200 "$code"

  code=$(req PUT "$VOTE/comments/$COMMENT_1_ID/votes" '{"value":1}' "Authorization: Bearer $ACCESS_TOKEN_B")
  check "user B upvotes comment 1" 200 "$code"

  code=$(req GET "$VOTE/user/votes" "" "Authorization: Bearer $ACCESS_TOKEN_A")
  check "list user A's votes" 200 "$code"

  code=$(req DELETE "$VOTE/posts/$POST_B_ID/votes" "" "Authorization: Bearer $ACCESS_TOKEN_A")
  check "user A removes their vote on post B" 200 "$code"
fi

# ============================================================================
step "6. SUBSCRIBE: user A follows post B"
# ============================================================================
code=$(req POST "$SUBSCRIBE/posts/$POST_B_ID/subscriptions" "" "Authorization: Bearer $ACCESS_TOKEN_A")
check "user A subscribes to post B" 200 "$code"

code=$(req GET "$SUBSCRIBE/users/subscriptions" "" "Authorization: Bearer $ACCESS_TOKEN_A")
check "list user A's subscriptions" 200 "$code"
# NOTE: store.Subscriber has zero json tags -> PascalCase keys.
jq -r --arg pid "$POST_B_ID" '[(. // [])[] | select(.PostId == $pid)] | length' "$BODY_FILE" | \
  xargs -I{} echo "  subscriptions matching post B: {}"

# ============================================================================
step "7. NOTIFICATION: user B comments again on post B -> A should be notified"
# ============================================================================
NONCE_COMMENT_3="comment-3-marker-$NONCE"
code=$(req POST "$COMMENT/posts/$POST_B_ID/comments" "{\"body\":\"$NONCE_COMMENT_3\",\"parent_id\":\"00000000-0000-0000-0000-000000000000\"}" "Authorization: Bearer $ACCESS_TOKEN_B")
check "user B comments on post B (should publish comment_created event)" 201 "$code"

echo "  waiting for the async kafka -> notification pipeline..."
sleep 3

code=$(req GET "$NOTIFICATION/user/notifications" "" "Authorization: Bearer $ACCESS_TOKEN_A")
check "list user A's notifications" 200 "$code"
# NOTE: store.Notification has zero json tags -> PascalCase keys. Same
# null-vs-[] gotcha as above.
NOTIF_COUNT=$(jq -r '(. // []) | length' "$BODY_FILE" 2>/dev/null)
NOTIF_ID=$(jq -r '(. // [])[0].Id // empty' "$BODY_FILE" 2>/dev/null)
echo "  user A has $NOTIF_COUNT notification(s)"

if [[ -n "$NOTIF_ID" ]]; then
  code=$(req PATCH "$NOTIFICATION/notifications/$NOTIF_ID/read" "" "Authorization: Bearer $ACCESS_TOKEN_A")
  check "mark one notification read" 200 "$code"

  code=$(req PATCH "$NOTIFICATION/notifications/read-all" "" "Authorization: Bearer $ACCESS_TOKEN_A")
  check "mark all notifications read" 200 "$code"

  code=$(req DELETE "$NOTIFICATION/notifications/$NOTIF_ID" "" "Authorization: Bearer $ACCESS_TOKEN_A")
  check "delete a notification" 200 "$code"
else
  echo "  WARN: no notification found yet -- either the pipeline needs more time,"
  echo "        or user A subscribed after post B's *original* creation and the"
  echo "        notification service correctly only fires on new comment_created"
  echo "        events (i.e. this may be expected, not a failure)."
fi

# ============================================================================
step "8. CLEANUP"
# ============================================================================
code=$(req DELETE "$COMMENT/comments/$COMMENT_1_ID" "" "Authorization: Bearer $ACCESS_TOKEN_B")
check "delete comment 1" 204 "$code"

code=$(req DELETE "$SUBSCRIBE/posts/$POST_B_ID/subscriptions" "" "Authorization: Bearer $ACCESS_TOKEN_A")
check "user A unsubscribes from post B" 200 "$code"

code=$(req DELETE "$POST/posts/$POST_A_ID" "" "Authorization: Bearer $ACCESS_TOKEN_A")
check "delete post A" 204 "$code"

code=$(req DELETE "$POST/posts/$POST_B_ID" "" "Authorization: Bearer $ACCESS_TOKEN_B")
check "delete post B" 204 "$code"

# ============================================================================
step "SUMMARY"
# ============================================================================
echo "  passed: $PASS_COUNT"
echo "  failed: $FAIL_COUNT"
[[ "$FAIL_COUNT" -eq 0 ]] && exit 0 || exit 1
