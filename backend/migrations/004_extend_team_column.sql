-- Extend team column to support longer team names
ALTER TABLE players ALTER COLUMN team TYPE VARCHAR(100);
