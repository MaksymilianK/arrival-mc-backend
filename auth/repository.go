package auth

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	db2 "github.com/maksymiliank/arrival-mc-backend/util/db"
	web2 "github.com/maksymiliank/arrival-mc-backend/util/web"
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

type perm struct {
	Server int
	Perm string
}

func (repoS) getAllWebRanks() ([]rankWithPerms, error) {
	rows, err := db2.Conn().Query(context.Background(), "SELECT * FROM get_ranks_with_web_perms()")
	if err != nil {
		return nil, err
	}

	ranks := make([]rankWithPerms, 0)
	for rows.Next() {
		var r rankWithPerms
		if err := rows.Scan(&r.ID, &r.Level, &r.Name, &r.DisplayName, &r.ChatFormat, &r.Perms); err != nil {
			return nil, err
		}
	}
	return ranks, nil
}

func (repoS) getAllPerms(rank int) (map[int][]string, error) {
	rows, err := db2.Conn().Query(context.Background(), "SELECT * FROM get_perms($1)", rank)
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

func (repoS) createRank(rank rankCreation) (int, error) {
	row := db2.Conn().QueryRow(
		context.Background(),
		"SELECT * FROM create_rank($1, $2, $3, $4, $5)",
		rank.Level, rank.Name, rank.DisplayName, rank.ChatFormat, permsMapToSlice(rank.Perms),
	)

	var ID int
	if err := row.Scan(&ID); err != nil {
		return 0, err
	}
	return ID, nil
}

func (repoS) removeRank(ID int) error {
	_, err := db2.Conn().Exec(context.Background(), "SELECT * FROM remove_rank($1)", ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db2.ErrNoDataFound {
			return web2.ErrNotFound
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

	_, err := db2.Conn().Exec(
		context.Background(),
		"SELECT * FROM modify_rank($1, $2, $3, $4, $5, $6, $7)",
		ID, level, name, displayName, chatFormat, permsMapToSlice(rank.RemPerms), permsMapToSlice(rank.AddedPerms),
	)
	return err
}

func (repoS) getPlayerCredentials(nick string) (playerCredentials, error) {
	row := db2.Conn().QueryRow(context.Background(), "SELECT * FROM get_auth_data($1)", nick)

	var data playerCredentials
	if err := row.Scan(&data.id, &data.passHash, &data.rank); err != nil {
		if err == pgx.ErrNoRows {
			return playerCredentials{}, web2.ErrNotFound
		} else {
			return playerCredentials{}, err
		}
	}
	return data, nil
}

func (p *perm) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded")
	}

	if err := (pgtype.CompositeFields{&p.Server, &p.Perm}).DecodeBinary(ci, src); err != nil {
		return err
	}
	return nil
}

func (p perm) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	server := pgtype.Int2{int16(p.Server), pgtype.Present}
	perm := pgtype.Text{p.Perm, pgtype.Present}
	return (pgtype.CompositeFields{&server, &perm}).EncodeBinary(ci, buf)
}

func permsMapToSlice(permsMap map[int][]string) []perm {
	perms := make([]perm, 0)
	if permsMap == nil {
		return perms
	}

	for s, sp := range permsMap {
		for _, p := range sp {
			perms = append(perms, perm{s, p})
		}
	}
	return perms
}