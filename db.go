package shield

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func getDB(ctx context.Context) (*pgxpool.Pool, error) {
	godotenv.Load() // consumer err
	dbURL := os.Getenv("DATABASE_URL")
	db, err := connectToDB(ctx, dbURL)
	if err != nil {
		return nil, nil
	}
	return db, nil
}

// dbDesctructor is a pgxpool.Pool implementing caddy.Destruct
type dbDestructor struct {
	*pgxpool.Pool
}

func (d dbDestructor) Destruct() error {
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
