-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
gen_random_uuid(),
NOW(),
NOW(),
$1,
$2
)
RETURNING id, created_at, updated_at, email;

-- name: GetUsers :many
SELECT id, created_at, updated_at, email, is_chirpy_red FROM users
ORDER BY id;

-- name: GetUser :one
SELECT id, created_at, updated_at, email, is_chirpy_red FROM users 
WHERE id = $1
LIMIT 1;

-- name: AuthUser :one
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red FROM users 
WHERE email = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at, email;

-- name: UpgradeUserToChirpyRed :exec
UPDATE users
SET is_chirpy_red = true, updated_at = NOW()
WHERE id = $1;
