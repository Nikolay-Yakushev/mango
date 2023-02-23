package dbase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	_ "github.com/lib/pq"
)
var schema = `
CREATE TABLE if not exists users (
	id       uuid,
    login    text unique,
    password text,
    email    text
);
`

type User struct {
	Id       uuid.UUID
	Login    string
	Password string
	Email    string
}

const name = "db connection"

type DbStorage struct{
	client  *sqlx.DB
	log     *zap.Logger
}


func New(ctx context.Context, log *zap.Logger) (*DbStorage, error) {
	pgInfo:= "host=127.0.0.1 port=5432 user=docker password=docker dbname=docker_db sslmode=disable"

	db, err := sqlx.Open("postgres", pgInfo)
	if err != nil {
		panic(err.Error())
    }
	if err := db.Ping(); err != nil {
		panic(err.Error())
	}
	db.MustExec(schema)
	newLog := log.Named(name)
	client := &DbStorage{
		client: db,
		log : newLog,
	}
	return client, nil


}