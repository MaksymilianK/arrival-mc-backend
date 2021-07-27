package ban

import (
	"context"
	"github.com/jackc/pgtype"
	"github.com/maksymiliank/arrival-mc-backend/util/db"
)

type Repo interface {
	getAll(req banReq) ([]*banMin, error)
	createOne(ban banCreation) (int, error)
	modifyOne(ID int, modder int, ban banModificationReq) (int, error)
	deleteOne(ID int, modder int, removalReason string) error
}

type repoS struct {}

func NewRepo() Repo {
	return repoS{}
}

func (repoS) getAll(req banReq) ([]*banMin, error) {
	server := pgtype.Int2{int16(req.server), pgtype.Present}
	if req.server == 0 {
		server.Status = pgtype.Null
	}

	recipient := pgtype.Varchar{req.recipient, pgtype.Present}
	if req.recipient == "" {
		recipient.Status = pgtype.Null
	}

	startFrom := pgtype.Timestamp{Time: req.startFrom, Status: pgtype.Present}
	if req.startFrom.IsZero() {
		startFrom.Status = pgtype.Null
	}

	startTo := pgtype.Timestamp{Time: req.startTo, Status: pgtype.Present}
	if req.startTo.IsZero() {
		startTo.Status = pgtype.Null
	}

	expirationFrom := pgtype.Timestamp{Time: req.expirationFrom, Status: pgtype.Present}
	if req.expirationFrom.IsZero() {
		expirationFrom.Status = pgtype.Null
	}

	expirationTo := pgtype.Timestamp{Time: req.expirationTo, Status: pgtype.Present}
	if req.expirationTo.IsZero() {
		expirationTo.Status = pgtype.Null
	}

	rows, err := db.Conn().Query(
		context.Background(),
		"SELECT * FROM get_bans($1, $2, $3, $4, $5, $6)",
		server, recipient, startFrom, startTo, expirationFrom, expirationTo,
	)
	if err != nil {
		return nil, err
	}

	bans := make([]*banMin, 0)
	for rows.Next() {
		var b banMin
		if err := rows.Scan(&b.ID, &b.Server, &b.Recipient.Nick, &b.Recipient.Rank, &b.Start, &b.Expiration, &b.OldType);
			err != nil {
			return nil, err
		}
		bans = append(bans, &b)
	}
	return bans, nil
}

func (repoS) createOne(ban banCreation) (int, error) {
	row := db.Conn().QueryRow(
		context.Background(),
		"SELECT * FROM create_ban($1, $2, $3, $4, $5)",
		ban.server, ban.recipient, ban.giver, ban.duration, ban.reason,
	)

	var ID int
	if err := row.Scan(&ID); err != nil {
		return 0, err
	}
	return ID, nil
}

func (repoS) modifyOne(ID int, modder int, ban banModificationReq) (int, error) {
	server := pgtype.Int2{int16(ban.Server), pgtype.Present}
	if ban.Server == 0 {
		server.Status = pgtype.Null
	}

	recipient := pgtype.Varchar{ban.Recipient, pgtype.Present}
	if ban.Recipient == "" {
		recipient.Status = pgtype.Null
	}

	duration := pgtype.Interval{Days: int32(ban.Duration.Hours()) / 24, Status: pgtype.Present}
	if ban.Duration == 0 {
		duration.Status = pgtype.Null
	}

	reason := pgtype.Varchar{ban.Reason, pgtype.Present}
	if ban.Reason == "" {
		reason.Status = pgtype.Null
	}

	row := db.Conn().QueryRow(
		context.Background(),
		"SELECT * FROM modify_ban($1, $2, $3, $4, $5, &6, &7)",
		ID, server, recipient, duration, reason, modder, ban.ModificationReason,
	)

	var newBanID int
	if err := row.Scan(&newBanID); err != nil {
		return 0, err
	}
	return ID, nil
}

func (repoS) deleteOne(ID int, modder int, removalReason string) error {
	_, err := db.Conn().Exec(
		context.Background(),
		"SELECT * FROM remove_ban($1, $2, $3)",
		ID, modder, removalReason,
	)
	return err
}
