package player

import (
	"context"
	"github.com/jackc/pgtype"
	"github.com/maksymiliank/arrival-mc-backend/util/db"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
)

type Repo interface {
	getAll(page web.PageReq, nick string) (web.PageRes, error)
	getOne(nick string) (Res, error)
}

type repoS struct{}

func NewRepo() Repo {
	return repoS{}
}

func (repoS) getAll(page web.PageReq, nick string) (web.PageRes, error) {
	n := pgtype.Varchar{nick, pgtype.Present}
	if nick == "" {
		n.Status = pgtype.Null
	}

	rows, err := db.Conn().Query(
		context.Background(),
		"SELECT * FROM get_players($1, $2, $3)",
		page.Page, page.Size, n,
	)
	if err != nil {
		return web.PageRes{}, err
	}

	var total int
	players := make([]Res, 0)

	rows.Next()
	if err := rows.Scan(&total, nil, nil); err != nil {
		return web.PageRes{}, err
	}

	for rows.Next() {
		var p Res
		if err := rows.Scan(nil, &p.Nick, &p.Rank); err != nil {
			return web.PageRes{}, err
		}
		players = append(players, p)
	}

	return web.PageRes{total, players}, nil
}

func (repoS) getOne(nick string) (Res, error) {
	row := db.Conn().QueryRow(context.Background(), "SELECT * FROM get_player($1)", nick)

	var p Res
	if err := row.Scan(&p.Nick, &p.Rank); err != nil {
		return Res{}, err
	}
	return p, nil
}
