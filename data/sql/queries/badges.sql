-- name: CreateBadge :one
INSERT INTO badges (
    token_hash,
    subject,
    status,
    color,
    style
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, token_hash, subject, status, color, style, created_at, updated_at;

-- name: GetBadgeByID :one
SELECT id, token_hash, subject, status, color, style, created_at, updated_at
FROM badges
WHERE id = $1;

-- name: UpdateBadge :one
UPDATE badges
SET subject = $2,
    status = $3,
    color = $4,
    style = $5,
    updated_at = now()
WHERE id = $1
RETURNING id, token_hash, subject, status, color, style, created_at, updated_at;

-- name: DeleteBadge :exec
DELETE FROM badges
WHERE id = $1;
