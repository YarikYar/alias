-- Add team_names column to rooms table
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS team_names JSONB DEFAULT '[]'::jsonb;

-- Add index for querying
CREATE INDEX IF NOT EXISTS idx_rooms_team_names ON rooms USING gin(team_names);
