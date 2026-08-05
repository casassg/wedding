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
-- Skips updates if invite has unsynced local RSVP changes.
INSERT INTO invites (
    invite_code, name, max_adults, max_kids,
    confirmed_adults, confirmed_kids, dietary_info, message_for_us, song_request, response_at,
    sheet_row, location,
    travel_bus_to, travel_pickup, travel_arrival_flight, travel_bus_return,
    travel_hotel, travel_notes, travel_cocktail, travel_brunch, travel_return_detail, travel_updated_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    datetime('now', 'utc')
)
ON CONFLICT(invite_code) DO UPDATE SET
    name                  = excluded.name,
    max_adults            = excluded.max_adults,
    max_kids              = excluded.max_kids,
    confirmed_adults      = excluded.confirmed_adults,
    confirmed_kids        = excluded.confirmed_kids,
    dietary_info          = excluded.dietary_info,
    message_for_us        = excluded.message_for_us,
    song_request          = excluded.song_request,
    response_at           = excluded.response_at,
    sheet_row             = excluded.sheet_row,
    location              = excluded.location,
    travel_bus_to         = excluded.travel_bus_to,
    travel_pickup         = excluded.travel_pickup,
    travel_arrival_flight = excluded.travel_arrival_flight,
    travel_bus_return     = excluded.travel_bus_return,
    travel_hotel          = excluded.travel_hotel,
    travel_notes          = excluded.travel_notes,
    travel_cocktail        = excluded.travel_cocktail,
    travel_brunch          = excluded.travel_brunch,
    travel_return_detail  = excluded.travel_return_detail,
    travel_updated_at     = excluded.travel_updated_at,
    updated_at            = excluded.updated_at
WHERE invites.response_at IS NULL OR invites.response_at <= invites.updated_at;
    -- Note: The WHERE clause prevents updates when local RSVP changes are pending,
    -- protecting local RSVP changes that haven't been pushed to the sheet yet.



-- name: UpdateTravelInfo :exec
-- Updates travel fields and bumps response_at so GetPendingSyncInvites picks it up.
UPDATE invites
SET
    travel_bus_to        = :input_bus_to,
    travel_pickup        = :input_pickup,
    travel_arrival_flight = :input_arrival_flight,
    travel_bus_return    = :input_bus_return,
    travel_hotel         = :input_hotel,
    travel_notes         = :input_notes,
    travel_cocktail      = :input_cocktail,
    travel_brunch        = :input_brunch,
    travel_return_detail = :input_return_detail,
    travel_updated_at    = datetime('now', 'utc'),
    response_at          = datetime('now', 'utc')
WHERE invite_code = :input_invite_code;

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
