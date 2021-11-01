package server

import (
	"context"
	"github.com/maksymiliank/arrival-mc-backend/util/db"
)

type Repo interface {
	getAll() ([]server, error)
}

type repoS struct{}

func NewRepo() Repo {
	return repoS{}
}

func (repoS) getAll() ([]server, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_servers()")
	if err != nil {
		return nil, err
	}

	servers := make([]server, 0)
	for rows.Next() {
		var s server
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, nil
}
