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
-- Master data (name, limits, location, sheet_row) always updates.
-- Guest-entered fields (RSVP, travel) are protected while pending sync.
-- datetime() around column refs normalizes RFC3339 vs space-separated timestamps
-- so the pending-changes guard compares correctly.
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
    -- Master data: always follow the sheet.
    name                  = excluded.name,
    max_adults            = excluded.max_adults,
    max_kids              = excluded.max_kids,
    sheet_row             = excluded.sheet_row,
    location              = excluded.location,
    -- Guest-entered fields: only overwrite when no local changes are pending.
    -- "pending" = response_at IS NOT NULL AND response_at > updated_at.
    -- datetime() normalizes both sides so RFC3339 and space-separated formats compare correctly.
    confirmed_adults      = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.confirmed_adults ELSE invites.confirmed_adults END,
    confirmed_kids        = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.confirmed_kids ELSE invites.confirmed_kids END,
    dietary_info          = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.dietary_info ELSE invites.dietary_info END,
    message_for_us        = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.message_for_us ELSE invites.message_for_us END,
    song_request          = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.song_request ELSE invites.song_request END,
    response_at           = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.response_at ELSE invites.response_at END,
    travel_bus_to         = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_bus_to ELSE invites.travel_bus_to END,
    travel_pickup         = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_pickup ELSE invites.travel_pickup END,
    travel_arrival_flight = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_arrival_flight ELSE invites.travel_arrival_flight END,
    travel_bus_return     = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_bus_return ELSE invites.travel_bus_return END,
    travel_hotel          = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_hotel ELSE invites.travel_hotel END,
    travel_notes          = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_notes ELSE invites.travel_notes END,
    travel_cocktail       = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_cocktail ELSE invites.travel_cocktail END,
    travel_brunch         = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_brunch ELSE invites.travel_brunch END,
    travel_return_detail  = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_return_detail ELSE invites.travel_return_detail END,
    travel_updated_at     = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN excluded.travel_updated_at ELSE invites.travel_updated_at END,
    updated_at            = CASE WHEN invites.response_at IS NULL OR datetime(invites.response_at) <= datetime(invites.updated_at)
                                 THEN datetime('now', 'utc') ELSE invites.updated_at END;



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
-- datetime() normalizes both sides so RFC3339 and space-separated formats compare correctly.
SELECT * FROM invites
WHERE response_at IS NOT NULL
  AND datetime(response_at) > datetime(updated_at)
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
