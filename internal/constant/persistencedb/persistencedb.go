package persistencedb

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"go.uber.org/zap"
)

type PersistenceDB struct {
	*db.Queries
	pool *pgxpool.Pool
	log  *zap.Logger
}

type Sibling string

func New(pool *pgxpool.Pool, log *zap.Logger) PersistenceDB {
	return PersistenceDB{
		Queries: db.New(pool),
		pool:    pool,
		log:     log,
	}
}
