-- name: CreateChirp :one
INSERT INTO chirps (id, user_id, created_at, updated_at, body)
VALUES (
    gen_random_uuid (),
    $1,
    NOW(),
    NOW(),
    $2
)
RETURNING *;

-- name: GetSingleChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at;

-- name: DeleteChirp :execresult
DELETE FROM chirps
WHERE id = $1 AND user_id = $2
RETURNING id, user_id;

-- name: DeleteChirps :exec
TRUNCATE TABLE chirps;
