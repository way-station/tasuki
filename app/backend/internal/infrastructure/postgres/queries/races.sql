-- name: GetRace :one
SELECT id, team_id, name, description, version
FROM races
WHERE id = $1;

-- name: SaveRace :execrows
-- Inserts a new race, or updates it only if the stored version is loaded_version.
-- No row is affected on a version conflict.
INSERT INTO races (id, team_id, name, description, version)
VALUES (@id, @team_id, @name, @description, @version)
ON CONFLICT (id) DO UPDATE
SET name        = EXCLUDED.name,
    description = EXCLUDED.description,
    version     = EXCLUDED.version,
    updated_at  = now()
WHERE races.version = @loaded_version;
