// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: deleteallchirps.sql

package database

import (
	"context"
)

const deleteAllChirps = `-- name: DeleteAllChirps :exec
DELETE FROM chirps
`

func (q *Queries) DeleteAllChirps(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteAllChirps)
	return err
}
