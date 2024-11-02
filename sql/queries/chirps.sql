
-- SQL Query to Create a Chirp in the Database
-- name: CreateChirp :one
INSERT INTO chirps (id, body, created_at, updated_at, user_id)
VALUES (
gen_random_uuid(),
$1,
NOW(),
NOW(),
$2
)
RETURNING *;


-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps 
WHERE id = $1
LIMIT 1;