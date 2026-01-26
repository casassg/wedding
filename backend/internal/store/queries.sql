-- name: GetInviteByInviteCode :one
SELECT * FROM invites WHERE invite_code = ?;

-- name: UpdateRSVP :exec
-- Updates RSVP details and forces a sync (synced_at = NULL).
UPDATE invites
SET
    confirmed_adults = :input_confirmed_adults,
    confirmed_kids   = :input_confirmed_kids,
    dietary_info     = :input_dietary_info,
    message_for_us   = :input_message,
    song_request     = :input_song,
    response_country = :input_response_country,
    response_at      = datetime('now', 'utc'),
    updated_at       = datetime('now', 'utc'),
    synced_at        = NULL
WHERE
    invite_code = :input_invite_code
    -- Validation Logic:
    AND :input_confirmed_adults <= max_adults
    AND :input_confirmed_kids   <= max_kids;

-- name: UpsertInvite :exec
-- Syncs Master Data from Google Sheets -> DB.
INSERT INTO invites (
    invite_code, name, max_adults, max_kids, confirmed_adults, sheet_row, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, datetime('now', 'utc')
)
ON CONFLICT(invite_code) DO UPDATE SET
    name       = excluded.name,
    max_adults = excluded.max_adults,
    max_kids   = excluded.max_kids,
    sheet_row  = excluded.sheet_row,
    confirmed_adults = excluded.confirmed_adults,
    updated_at = excluded.updated_at;
    -- Note: We still rely on existing synced_at value (no change needed)
    -- because DO UPDATE does not touch columns unless specified.

-- name: DeleteInvite :exec
-- HARD DELETE: This permanently removes the row.
DELETE FROM invites
WHERE invite_code = ?;

-- name: GetPendingSyncInvites :many
-- Finds rows that have responded but haven't been synced OR have changed since sync.
SELECT * FROM invites
WHERE response_at IS NOT NULL
  AND (synced_at IS NULL OR response_at > synced_at)
ORDER BY response_at ASC;

-- name: MarkInviteSynced :exec
UPDATE invites
SET
    synced_at = datetime('now', 'utc'),
    updated_at = datetime('now', 'utc')
WHERE invite_code = ?;
