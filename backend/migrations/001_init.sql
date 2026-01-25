-- Initial schema for wedding RSVP system
CREATE TABLE IF NOT EXISTS invites (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid            TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    max_adults      INTEGER NOT NULL DEFAULT 1,
    max_kids        INTEGER NOT NULL DEFAULT 0,
    
    -- RSVP response fields
    attending       BOOLEAN,
    adult_count     INTEGER,
    kid_count       INTEGER,
    dietary_info    TEXT,
    transport_needs TEXT,
    
    -- Metadata
    response_at     DATETIME,
    response_country TEXT,
    
    -- Sync tracking
    sheet_row       INTEGER,
    synced_at       DATETIME,
    deleted_at      DATETIME,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invites_uuid ON invites(uuid);
CREATE INDEX IF NOT EXISTS idx_invites_deleted ON invites(deleted_at);
CREATE INDEX IF NOT EXISTS idx_invites_synced ON invites(synced_at);
