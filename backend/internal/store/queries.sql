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
    response_at      = datetime('now', 'utc')-- Mark as needing sync
WHERE
    invite_code = :input_invite_code
    -- Validation Logic:
    AND :input_confirmed_adults <= max_adults
    AND :input_confirmed_kids   <= max_kids;

-- name: UpsertInvite :exec
-- Syncs Master Data from Google Sheets -> DB.
-- Skips updates if invite has unsynced local changes (synced_at IS NULL).
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
    updated_at = excluded.updated_at
WHERE invites.response_at IS NULL OR invites.response_at <= invites.updated_at;
    -- Note: The WHERE clause prevents updates when synced_at IS NULL,
    -- protecting local RSVP changes that haven't been pushed to the sheet yet.



-- name: DeleteInvite :exec
-- HARD DELETE: This permanently removes the row.
DELETE FROM invites
WHERE invite_code = ?;

-- name: GetPendingSyncInvites :many
-- Finds rows that have responded but haven't been synced OR have changed since sync.
SELECT * FROM invites
WHERE response_at IS NOT NULL
  AND response_at > updated_at
ORDER BY response_at ASC;

-- name: MarkInviteSynced :exec
UPDATE invites
SET
    updated_at = datetime('now', 'utc')
WHERE invite_code = ?;

-- =====================
-- Schedule Events Queries
-- =====================

-- name: GetScheduleEvents :many
-- Returns all schedule events ordered by start time.
-- Only public events are stored in the DB (filtered during sync).
SELECT * FROM schedule_events
ORDER BY start_time ASC;

-- name: DeleteAllScheduleEvents :exec
-- Clears all schedule events before a full re-sync from sheet.
DELETE FROM schedule_events;

-- name: InsertScheduleEvent :exec
-- Inserts a single schedule event during sync.
INSERT INTO schedule_events (
    start_time, end_time,
    event_name_es, event_name_en, event_name_ca,
    location,
    description_es, description_en, description_ca,
    updated_at
) VALUES (
    ?, ?,
    ?, ?, ?,
    ?,
    ?, ?, ?,
    datetime('now', 'utc')
);
