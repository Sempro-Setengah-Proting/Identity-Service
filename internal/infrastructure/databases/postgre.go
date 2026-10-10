package databases

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDb(dbURL string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Connected to database successfully!")
	return db, nil
}
