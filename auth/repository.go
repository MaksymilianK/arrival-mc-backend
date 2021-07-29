package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	db "github.com/maksymiliank/arrival-mc-backend/util/db"
	web "github.com/maksymiliank/arrival-mc-backend/util/web"
	"strings"
)

type Repo interface {
	getAllWebRanks() ([]rankWithPerms, error)
	getAllPerms(rank int) (map[int][]string, error)
	createRank(rank rankCreation) (int, error)
	removeRank(ID int) error
	modifyRank(ID int, rank rankModification) error

	getPlayerCredentials(nick string) (playerCredentials, error)
}

type repoS struct {}

func NewRepo() Repo {
	return repoS{}
}

type permission struct {
	server int
	perm string
}

func (repoS) getAllWebRanks() ([]rankWithPerms, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_ranks_with_web_perms()")
	if err != nil {
		return nil, err
	}

	ranks := make([]rankWithPerms, 0)
	for rows.Next() {
		var r rankWithPerms
		if err := rows.Scan(&r.ID, &r.Level, &r.Name, &r.DisplayName, &r.ChatFormat, &r.Perms); err != nil {
			return nil, err
		}
		ranks = append(ranks, r)
	}
	return ranks, nil
}

func (repoS) getAllPerms(rank int) (map[int][]string, error) {
	rows, err := db.Conn().Query(context.Background(), "SELECT * FROM get_perms($1)", rank)
	if err != nil {
		return nil, err
	}

	perms := make(map[int][]string)
	for rows.Next() {
		var server int
		var perm string
		if err := rows.Scan(&server, &perm); err != nil {
			return nil, err
		}

		if _, ok := perms[server]; !ok {
			perms[server] = make([]string, 0)
		}
		perms[server] = append(perms[server], perm)
	}
	return perms, nil
}

func pfs(permsMap map[int][]string) string {
	var perms strings.Builder
	perms.WriteString("ARRAY[")
	for _, p := range permsMapToSlice(permsMap) {
		perms.WriteString(fmt.Sprintf("(%d,'%s'),", p.server, p.perm))
	}
	p := perms.String()
	if p[len(p) - 1] != '[' {
		p = p[:len(p) - 1]
	}
	p += "]::permission[]"
	return p
}

func (repoS) createRank(rank rankCreation) (int, error) {
	row := db.Conn().QueryRow(
		context.Background(),fmt.Sprintf(
		"SELECT * FROM create_rank($1, $2, $3, $4, %s)", pfs(rank.Perms)),
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
		context.Background(), fmt.Sprintf(
		"SELECT * FROM modify_rank($1, $2, $3, $4, $5, %s, %s)", pfs(rank.RemPerms), pfs(rank.AddedPerms)),
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

	if err := (pgtype.CompositeFields{&p.server, &p.perm}).DecodeBinary(ci, src); err != nil {
		return err
	}
	return nil
}

func (p permission) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	server := pgtype.Int2{int16(p.server), pgtype.Present}
	perm := pgtype.Varchar{p.perm, pgtype.Present}
	return (pgtype.CompositeFields{&server, &perm}).EncodeBinary(ci, buf)
}

func permsMapToSlice(permsMap map[int][]string) []permission {
	perms := make([]permission, 0)
	if permsMap == nil {
		return perms
	}

	for s, sp := range permsMap {
		for _, p := range sp {
			perms = append(perms, permission{s, p})
		}
	}
	return perms
}