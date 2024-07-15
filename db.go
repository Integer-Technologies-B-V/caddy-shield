package shield

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connectToDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		return nil, err
	}
	return dbPool, nil
}

type dbDestructor struct {
	db *pgxpool.Pool
}

func (d dbDestructor) Destruct() error {
	d.db.Close()
	return nil
}
