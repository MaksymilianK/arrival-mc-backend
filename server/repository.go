package server

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

type repoI interface {
	setUp(db *pgxpool.Pool)
	existsByID(ID int) bool
	getAll() []Server
}

type repoS struct {
	all []Server
	byID map[int]Server
}

var repo repoI = &repoS{
	make([]Server, 0),
	make(map[int]Server),
}

func (r *repoS) setUp(db *pgxpool.Pool) {
	rows, err := db.Query(context.Background(), "SELECT * FROM get_servers()")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var s Server
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			panic(err)
		}
		r.all = append(r.all, s)
		r.byID[s.ID] = s
	}
}

func (r *repoS) existsByID(ID int) bool {
	_, ok := r.byID[ID]
	return ok
}

func (r *repoS) getAll() []Server {
	return r.all
}
