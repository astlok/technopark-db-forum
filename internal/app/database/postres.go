package database

import (
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
)

type Postgres struct {
	db *pgx.ConnPool
}

func NewPostgres() (*Postgres, error) {
	conf := pgx.ConnConfig{
		User:                 "postgres",
		Database:             "postgres",
		Password:             "admin",
		PreferSimpleProtocol: false,
	}

	poolConf := pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 100,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	}
	db, err := pgx.NewConnPool(poolConf)
	if err != nil {
		return nil, err
	}
	return &Postgres{
		db: db,
	}, nil
}

func (p *Postgres) GetPostgres() *pgx.ConnPool {
	return p.db
}

func (p *Postgres) Close() error {
	p.db.Close()
	return nil
}
