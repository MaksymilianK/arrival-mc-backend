package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/maksymiliank/arrival-mc-backend/util/db"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"log"
)

type Repo interface {
	getRanks() ([]*rankFull, error)
	getServerRanks(serv int) ([]*rankWithPerms, error)
	getAllPerms(rank int) (map[int][]string, map[int][]string, error)
	createRank(rank rankCreation) (int, error)
	removeRank(ID int) error
	modifyRank(ID int, rank rankModification) error

	getPlayerCredentials(nick string) (playerCredentials, error)
}

type repoS struct{}

func SetUpRepo() Repo {
	return repoS{}
}

func (repoS) getRanks() ([]*rankFull, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_ranks()")
	if err != nil {
		return nil, err
	}

	ranks := make([]*rankFull, 0)
	for rows.Next() {
		r := rankFull{Perms: make(map[int][]string), NegatedPerms: make(map[int][]string)}
		var perms []db.Perm
		if err := rows.Scan(&r.ID, &r.Level, &r.Name, &r.DisplayName, &r.ChatFormat, &perms); err != nil {
			return nil, err
		}

		for _, p := range perms {
			if p.Negated {
				if _, ok := r.NegatedPerms[p.Server]; !ok {
					r.NegatedPerms[p.Server] = make([]string, 0)
				}
				r.NegatedPerms[p.Server] = append(r.NegatedPerms[p.Server], p.Value)
			} else {
				if _, ok := r.Perms[p.Server]; !ok {
					r.Perms[p.Server] = make([]string, 0)
				}
				r.Perms[p.Server] = append(r.Perms[p.Server], p.Value)
			}
		}

		ranks = append(ranks, &r)
	}

	return ranks, nil
}

func (repoS) getServerRanks(serv int) ([]*rankWithPerms, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_server_ranks($1)", serv)
	if err != nil {
		return nil, err
	}

	ranks := make([]*rankWithPerms, 0)
	for rows.Next() {
		var r rankWithPerms
		if err := rows.Scan(&r.ID, &r.Level, &r.Name, &r.DisplayName, &r.ChatFormat, &r.Perms, &r.NegatedPerms); err != nil {
			return nil, err
		}
		ranks = append(ranks, &r)
	}
	return ranks, nil
}

func (repoS) getAllPerms(rank int) (map[int][]string, map[int][]string, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_permissions($1)", rank)
	if err != nil {
		return nil, nil, err
	}

	perms := make(map[int][]string)
	negatedPerms := make(map[int][]string)

	for rows.Next() {
		var p db.Perm

		if err := rows.Scan(&p.Server, &p.Value, &p.Negated); err != nil {
			return nil, nil, err
		}

		if _, ok := perms[p.Server]; !ok {
			perms[p.Server] = make([]string, 0)
			negatedPerms[p.Server] = make([]string, 0)
		}

		if p.Negated {
			negatedPerms[p.Server] = append(negatedPerms[p.Server], p.Value)
		} else {
			perms[p.Server] = append(perms[p.Server], p.Value)
		}
	}
	return perms, negatedPerms, nil
}

func (repoS) createRank(rank rankCreation) (int, error) {
	perms := make(db.PermArray, 0)
	for s, sp := range rank.Perms {
		for _, p := range sp {
			perms = append(perms, db.Perm{s, p, false, pgtype.Present})
		}
	}
	for s, sp := range rank.NegatedPerms {
		for _, p := range sp {
			perms = append(perms, db.Perm{s, p, true, pgtype.Present})
		}
	}

	log.Println(perms)

	row := db.Conn().QueryRow(
		context.Background(),
		fmt.Sprintf(
			"SELECT * FROM create_rank($1, $2, $3, $4, $5)",
		),
		rank.Level, rank.Name, rank.DisplayName, rank.ChatFormat, perms,
	)

	var ID int
	if err := row.Scan(&ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Message) // => syntax error at end of input
			fmt.Println(pgErr.Code) // => 42601
		}
		return 0, err
	}
	return ID, nil
}

func (repoS) removeRank(ID int) error {
	_, err := db.Conn().Exec(context.Background(), "SELECT * FROM remove_rank($1)", ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.ErrNoDataFound {
			return web.ErrNotFound
		}
	}
	return err
}

func (repoS) modifyRank(ID int, rank rankModification) error {
	level := pgtype.Int2{int16(rank.Level), pgtype.Present}
	if rank.Level == 0 {
		level.Status = pgtype.Null
	}

	name := pgtype.Text{rank.Name, pgtype.Present}
	if rank.Name == "" {
		name.Status = pgtype.Null
	}

	displayName := pgtype.Text{rank.DisplayName, pgtype.Present}
	if rank.DisplayName == "" {
		displayName.Status = pgtype.Null
	}

	chatFormat := pgtype.Text{rank.ChatFormat, pgtype.Present}
	if rank.ChatFormat == "" {
		chatFormat.Status = pgtype.Null
	}

	_, err := db.Conn().Exec(
		context.Background(),
		fmt.Sprintf("SELECT * FROM modify_rank($1, $2, $3, $4, $5, $6, $7)"),
		ID, level, name, displayName, chatFormat, rank.AddedNegatedPerms, rank.RemovedNegatedPerms,
	)
	return err
}

func (repoS) getPlayerCredentials(nick string) (playerCredentials, error) {
	row := db.Conn().QueryRow(context.Background(), "SELECT * FROM get_auth_data($1)", nick)

	var data playerCredentials
	if err := row.Scan(&data.id, &data.passHash, &data.rank); err != nil {
		if err == pgx.ErrNoRows {
			return playerCredentials{}, web.ErrNotFound
		} else {
			return playerCredentials{}, err
		}
	}
	return data, nil
}
