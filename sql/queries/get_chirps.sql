-- name: GetChirps :many
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE sqlc.arg(include_all)::boolean OR user_id = sqlc.arg(author_id)::uuid
ORDER BY created_at; 