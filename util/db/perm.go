package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgtype"
)

type Perm struct {
	Server  int
	Value   string
	Negated bool
	Status pgtype.Status
}

type PermArray []Perm

var permOID uint32

func setUpPermOID() {
	row := conn.QueryRow(context.Background(), "SELECT oid FROM pg_type WHERE typname = $1", "permission")

	if err := row.Scan(&permOID); err != nil {
		panic(err)
	}
}

func (p *Perm) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded")
	}

	if err := (pgtype.CompositeFields{&p.Server, &p.Value, &p.Negated}).DecodeBinary(ci, src); err != nil {
		return err
	}
	return nil
}

func (p Perm) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	return (pgtype.CompositeFields{
		&pgtype.Int2{int16(p.Server), pgtype.Present},
		&pgtype.Varchar{p.Value, pgtype.Present},
		&pgtype.Bool{p.Negated, pgtype.Present},
	}).EncodeBinary(ci, buf)
}

func (p *Perm) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded")
	}

	if err := (pgtype.CompositeFields{&p.Server, &p.Value, &p.Negated}).DecodeText(ci, src); err != nil {
		return err
	}
	return nil
}

func (p Perm) EncodeText(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	return (pgtype.CompositeFields{
		&pgtype.Int2{int16(p.Server), pgtype.Present},
		&pgtype.Varchar{p.Value, pgtype.Present},
		&pgtype.Bool{p.Negated, pgtype.Present},
	}).EncodeText(ci, buf)
}

func (p *Perm) Set(src interface{}) error {
	switch t := src.(type) {
	case Perm:
		p.Status = pgtype.Present
		p.Server = t.Server
		p.Value = t.Value
		p.Negated = t.Negated
	case *Perm:
		if t == nil {
			return errors.New("cannot set nil Perm")
		} else {
			return p.Set(*t)
		}
	default:
		return errors.New("cannot set non-Perm type as Perm")
	}
	return nil
}

func (p *Perm) Get() interface{} {
	switch p.Status {
	case pgtype.Present:
		return p
	case pgtype.Null:
		return nil
	default:
		return p.Status
	}
}

func (p *Perm) AssignTo(dst interface{}) error {
	return fmt.Errorf("cannot assign %v to %T", p, dst)
}

func (*PermArray) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded")
	}

	if err := permArrType().DecodeBinary(ci, src); err != nil {
		return err
	}
	return nil
}

func (pa PermArray) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	arr := permArrType()
	err = arr.Set(pa)
	if err != nil {
		panic(err)
	}

	return arr.EncodeBinary(ci, buf)
}

func permArrType() *pgtype.ArrayType {
	return pgtype.NewArrayType("permissions", permOID, func() pgtype.ValueTranscoder {
		return &Perm{}
	})
}
