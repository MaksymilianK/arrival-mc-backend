package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maksymiliank/arrival-mc-backend/util"
)

type config struct {
	host string
	port int
	database string
	user string
	password string
}

var ErrPersistence = errors.New("error while querying db")
var conn *pgxpool.Pool

func SetUp(cfg util.DBConfig) {
	c, err := pgxpool.Connect(
		context.Background(),
		fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database),
	)
	if err != nil {
		panic(err)
	}
	conn = c
}

func Conn() *pgxpool.Pool {
	return conn
}