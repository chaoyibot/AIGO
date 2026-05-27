-- +migrate Up
BEGIN;

ALTER TABLE broadcasts ADD COLUMN IF NOT EXISTS is_pinned BOOLEAN DEFAULT false;
ALTER TABLE broadcasts ADD COLUMN IF NOT EXISTS pinned_at TIMESTAMPTZ;
ALTER TABLE broadcasts ADD COLUMN IF NOT EXISTS pin_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_broadcasts_pinned ON broadcasts(is_pinned) WHERE is_pinned = true;
CREATE INDEX IF NOT EXISTS idx_broadcasts_pin_expires ON broadcasts(pin_expires_at);

COMMIT;

-- +migrate Down
BEGIN;

DROP INDEX IF EXISTS idx_broadcasts_pin_expires;
DROP INDEX IF EXISTS idx_broadcasts_pinned;
ALTER TABLE broadcasts DROP COLUMN IF EXISTS pin_expires_at;
ALTER TABLE broadcasts DROP COLUMN IF EXISTS pinned_at;
ALTER TABLE broadcasts DROP COLUMN IF EXISTS is_pinned;

COMMIT;