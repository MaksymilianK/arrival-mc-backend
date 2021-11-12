package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maksymiliank/arrival-mc-backend/util"
	"log"
	"os"
)

const setupFile = "./db-setup.sql"

var ErrPersistence = errors.New("error while querying db")
var conn *pgxpool.Pool

func SetUp(cfg util.DBConfig, fullSetup bool) {
	log.Println(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database))
	c, err := pgxpool.Connect(
		context.Background(),
		fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database),
	)
	if err != nil {
		panic(err)
	}
	conn = c

	if fullSetup {
		setUpFullDB()
	}

	setUpPermOID()
}

func setUpFullDB() {
	if _, err := conn.Exec(context.Background(), readSetupFile()); err != nil {
		panic(err)
	}
}

func readSetupFile() string {
	data, err := os.ReadFile(setupFile)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func Conn() *pgxpool.Pool {
	return conn
}
