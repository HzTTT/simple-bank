-- name: CreateAccount :one
INSERT INTO accounts (
    owner,
    balance,
    currency
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetAccount :one
SELECT *
FROM accounts 
WHERE id = $1 LIMIT 1;

-- name: ListAccount :one
SELECT *
FROM accounts
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $1
WHERE id = $2;
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM account
WHERE id = $1;

