-- Add support for multiple teams

-- Add num_teams column to rooms
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS num_teams INT NOT NULL DEFAULT 2;

-- Add category column to rooms if not exists
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS category VARCHAR(50) NOT NULL DEFAULT 'general';

-- Update team column in players to support longer team names
ALTER TABLE players ALTER COLUMN team TYPE VARCHAR(20);

-- Add category column to words if not exists
ALTER TABLE words ADD COLUMN IF NOT EXISTS category VARCHAR(50) NOT NULL DEFAULT 'general';

-- Create index on category
CREATE INDEX IF NOT EXISTS idx_words_category ON words(category);
CREATE INDEX IF NOT EXISTS idx_rooms_category ON rooms(category);
