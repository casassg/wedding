CREATE TABLE IF NOT EXISTS invites (
    invite_code TEXT PRIMARY KEY,
    name TEXT NOT NULL,

    -- Constraints to prevent negative numbers
    max_adults INTEGER NOT NULL DEFAULT 1 CHECK (max_adults >= 0),
    max_kids INTEGER NOT NULL DEFAULT 0 CHECK (max_kids >= 0),
    confirmed_adults INTEGER NOT NULL DEFAULT 0 CHECK (confirmed_adults >= 0),
    confirmed_kids INTEGER NOT NULL DEFAULT 0 CHECK (confirmed_kids >= 0),

    dietary_info TEXT NOT NULL DEFAULT '',
    message_for_us TEXT NOT NULL DEFAULT '',
    song_request TEXT NOT NULL DEFAULT '',

    response_at DATETIME,
    sheet_row INTEGER,

    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc'))
);

-- OPTIMIZATION: Index for the Sync Queue
-- We index 'response_at' because we filter by it (IS NOT NULL) and sort by it.
-- This makes finding the next items to sync very fast.
CREATE INDEX IF NOT EXISTS idx_invites_response_at
ON invites(response_at)
WHERE response_at IS NOT NULL;
