-- name: GetInviteByUUID :one
SELECT * FROM invites
WHERE uuid = ? AND deleted_at IS NULL;

-- name: UpdateRSVP :exec
-- Update RSVP response and reset synced_at to trigger sync to Google Sheets
-- The CASE statement ensures synced_at is set to NULL only when response data changes
UPDATE invites
SET attending = ?,
    adult_count = ?,
    kid_count = ?,
    dietary_info = ?,
    message_for_us = ?,
    song_request = ?,
    response_at = ?,
    response_country = ?,
    synced_at = NULL,  -- Reset to trigger sync to sheet
    updated_at = ?
WHERE uuid = ? AND deleted_at IS NULL;

-- name: UpsertInvite :exec
-- Upsert invite data from Google Sheets
-- Only updates master data (name, max_adults, max_kids, sheet_row)
-- Preserves RSVP data and synced_at timestamp if record already exists
-- On INSERT: sets synced_at to the provided timestamp
-- On UPDATE: preserves existing synced_at (even if NULL) to maintain RSVP sync state
INSERT INTO invites (uuid, name, max_adults, max_kids, sheet_row, synced_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uuid) DO UPDATE SET
    name = excluded.name,
    max_adults = excluded.max_adults,
    max_kids = excluded.max_kids,
    sheet_row = excluded.sheet_row,
    -- Preserve existing synced_at on UPDATE (don't overwrite with new timestamp)
    synced_at = invites.synced_at,
    updated_at = excluded.updated_at,
    deleted_at = NULL;

-- name: SoftDeleteInvite :exec
UPDATE invites
SET deleted_at = ?, updated_at = ?
WHERE uuid = ? AND deleted_at IS NULL;

-- name: GetPendingSyncInvites :many
SELECT * FROM invites
WHERE deleted_at IS NULL
  AND response_at IS NOT NULL
  AND (synced_at IS NULL OR response_at > synced_at)
ORDER BY response_at ASC;

-- name: MarkInviteSynced :exec
UPDATE invites
SET synced_at = ?, updated_at = ?
WHERE uuid = ?;
