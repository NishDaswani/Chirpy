-- name: GetUserByEmail :one
SELECT ID, created_at, updated_at, email, hashed_password, is_chirpy_red
FROM users
WHERE email = $1;