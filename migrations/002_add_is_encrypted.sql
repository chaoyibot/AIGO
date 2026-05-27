BEGIN;

-- Add is_encrypted column for end-to-end encrypted messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS is_encrypted BOOLEAN DEFAULT false;

COMMIT;
