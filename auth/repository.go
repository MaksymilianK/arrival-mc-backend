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
	"strings"
)

type Repo interface {
	getRanks() ([]*rankFull, error)
	getAllWebRanks() ([]rankWithPerms, error)
	getAllPerms(rank int) (map[int][]string, map[int][]string, error)
	createRank(rank rankCreation) (int, error)
	removeRank(ID int) error
	modifyRank(ID int, rank rankModification) error

	getPlayerCredentials(nick string) (playerCredentials, error)
}

type repoS struct{}

func NewRepo() Repo {
	return repoS{}
}

type permission struct {
	Server int
	Perm   string
	Negated bool
}

func (repoS) getRanks() ([]*rankFull, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_ranks()")
	if err != nil {
		return nil, err
	}

	ranks := make([]*rankFull, 0)
	for rows.Next() {
		r := rankFull{Perms: make(map[int][]string), NegatedPerms: make(map[int][]string)}
		var perms []permission
		if err := rows.Scan(&r.ID, &r.Level, &r.Name, &r.DisplayName, &r.ChatFormat, &perms); err != nil {
			return nil, err
		}

		for _, p := range perms {
			if p.Negated {
				if _, ok := r.NegatedPerms[p.Server]; !ok {
					r.NegatedPerms[p.Server] = make([]string, 0)
				}
				r.NegatedPerms[p.Server] = append(r.NegatedPerms[p.Server], p.Perm)
			} else {
				if _, ok := r.Perms[p.Server]; !ok {
					r.Perms[p.Server] = make([]string, 0)
				}
				r.Perms[p.Server] = append(r.Perms[p.Server], p.Perm)
			}
		}

		ranks = append(ranks, &r)
	}

	return ranks, nil
}

func (repoS) getAllWebRanks() ([]rankWithPerms, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_ranks_with_web_perms()")
	if err != nil {
		return nil, err
	}

	ranks := make([]rankWithPerms, 0)
	for rows.Next() {
		var r rankWithPerms
		if err := rows.Scan(&r.ID, &r.Level, &r.Name, &r.DisplayName, &r.ChatFormat, &r.Perms, &r.NegatedPerms); err != nil {
			return nil, err
		}
		ranks = append(ranks, r)
	}
	return ranks, nil
}

func (repoS) getAllPerms(rank int) (map[int][]string, map[int][]string, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_perms($1)", rank)
	if err != nil {
		return nil, nil, err
	}

	perms := make(map[int][]string)
	negatedPerms := make(map[int][]string)

	for rows.Next() {
		var p permission

		if err := rows.Scan(&p.Server, &p.Perm, &p.Negated); err != nil {
			return nil, nil, err
		}

		if _, ok := perms[p.Server]; !ok {
			perms[p.Server] = make([]string, 0)
			negatedPerms[p.Server] = make([]string, 0)
		}

		if p.Negated {
			negatedPerms[p.Server] = append(negatedPerms[p.Server], p.Perm)
		} else {
			perms[p.Server] = append(perms[p.Server], p.Perm)
		}
	}
	return perms, negatedPerms, nil
}

func pfs(permSlice []permission) string {
	var perms strings.Builder
	perms.WriteString("ARRAY[")
	for _, p := range permSlice {
		perms.WriteString(fmt.Sprintf("(%d,'%s', %t),", p.Server, p.Perm, p.Negated))
	}
	p := perms.String()
	if p[len(p)-1] != '[' {
		p = p[:len(p)-1]
	}
	p += "]::permission[]"
	return p
}

func (repoS) createRank(rank rankCreation) (int, error) {
	row := db.Conn().QueryRow(
		context.Background(),
		fmt.Sprintf(
			"SELECT * FROM create_rank($1, $2, $3, $4, %s)",
			pfs(bothPermsMapToSlice(rank.Perms, rank.NegatedPerms)),
		),
		rank.Level, rank.Name, rank.DisplayName, rank.ChatFormat,
	)

	var ID int
	if err := row.Scan(&ID); err != nil {
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
		fmt.Sprintf(
			"SELECT * FROM modify_rank($1, $2, $3, $4, $5, %s, %s)",
			pfs(bothPermsMapToSlice(rank.AddedPerms, rank.AddedNegatedPerms)),
			pfs(bothPermsMapToSlice(rank.RemovedPerms, rank.RemovedNegatedPerms)),
		),
		ID, level, name, displayName, chatFormat,
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

func (p *permission) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded")
	}

	if err := (pgtype.CompositeFields{&p.Server, &p.Perm, &p.Negated}).DecodeBinary(ci, src); err != nil {
		return err
	}
	return nil
}

func (p permission) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	server := pgtype.Int2{int16(p.Server), pgtype.Present}
	perm := pgtype.Varchar{p.Perm, pgtype.Present}
	negated := pgtype.Bool{p.Negated, pgtype.Present}
	return (pgtype.CompositeFields{&server, &perm, &negated}).EncodeBinary(ci, buf)
}

func permsMapToSlice(permsMap map[int][]string, negated bool) []permission {
	if negated {
		return bothPermsMapToSlice(make(map[int][]string), permsMap)
	} else {
		return bothPermsMapToSlice(permsMap, make(map[int][]string))
	}
}

func bothPermsMapToSlice(permsMap map[int][]string, negatedPermsMap map[int][]string) []permission {
	perms := make([]permission, 0)

	if permsMap != nil {
		for s, sp := range permsMap {
			for _, p := range sp {
				perms = append(perms, permission{s, p, false})
			}
		}
	}

	if negatedPermsMap != nil {
		for s, sp := range negatedPermsMap {
			for _, p := range sp {
				perms = append(perms, permission{s, p, true})
			}
		}
	}

	return perms
}
