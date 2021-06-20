package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type config struct {
	host string
	port int
	database string
	user string
	password string
}

var conn *pgxpool.Pool

func init() {
	cfg := readCfg()
	c, err := pgxpool.Connect(
		context.Background(),
		fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.user, cfg.password, cfg.host, cfg.port, cfg.database),
	)
	if err != nil {
		panic(err)
	}
	conn = c
}

// Conn returns a database connection pool shared across application.
func Conn() *pgxpool.Pool {
	return conn
}

func readCfg() config {
	data, err := ioutil.ReadFile("./db-config.yaml")
	if err != nil {
		panic(err)
	}

	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
