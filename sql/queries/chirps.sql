
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

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;

-- name: GetChirpsByAuthor :many
SELECT id, body, created_at, updated_at, user_id
FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;


-- name: GetChirpsWithFilterAndSort :many
SELECT id, body, created_at, updated_at, user_id
FROM chirps
WHERE ($1 = '00000000-0000-0000-0000-000000000000' OR user_id = $1::uuid)
ORDER BY 
created_at $2;