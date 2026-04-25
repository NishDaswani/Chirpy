-- name: DeleteChirpByID :execrows
DELETE FROM chirps
WHERE id = $1;