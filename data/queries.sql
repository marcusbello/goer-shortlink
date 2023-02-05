-- -- name: FetchURLByID :one
SELECT * FROM links
WHERE id = $1 LIMIT 1;
--
-- -- name: FetchURLByCode :one
SELECT * FROM links
WHERE short = $1 LIMIT 1;

-- name: CreateShortLink :one
INSERT INTO links (
 short, url )
 VALUES ( $1, $2 )

