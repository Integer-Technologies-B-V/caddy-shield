package shield

import (
	"context"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func constructDB() (caddy.Destructor, error) {
	dbURL := os.Getenv("DB_URL")
	db, err := connectToDB(context.Background(), dbURL)
	if err != nil {
		return nil, nil
	}
	return &dbDestructor{Pool: db}, nil
}

// dbDesctructor is a dbpool and implements the caddy.Destruct interface
type dbDestructor struct {
	*pgxpool.Pool
}

func (d *dbDestructor) Destruct() error {
	d.Close()
	return nil
}

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
