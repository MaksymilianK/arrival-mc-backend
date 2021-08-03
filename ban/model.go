package ban

import (
	"github.com/maksymiliank/arrival-mc-backend/player"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"time"
)

type banMinModel struct {
	id         int
	server     int
	recipient  player.Res
	start      time.Time
	expiration time.Time
	oldType    rune
}

type banMinRes struct {
	ID         int        `json:"id"`
	Server     int        `json:"server"`
	Recipient  player.Res `json:"recipient"`
	Start      int64      `json:"start"`
	Expiration int64      `json:"expiration"`
	OldType    rune       `json:"oldType;omitempty"`
}

type banFullModel struct {
	banMinModel
	actualExpiration   time.Time
	giver              player.Res
	reason             string
	newBan             int
	modder             player.Res
	modificationReason string
}

type banFullRes struct {
	banMinRes
	ActualExpiration   int64      `json:"actualExpiration;omitempty"`
	Giver              player.Res `json:"giver"`
	Reason             string     `json:"reason"`
	NewBan             int        `json:"newBan;omitempty"`
	Modder             player.Res `json:"modder;omitempty"`
	ModificationReason string     `json:"modificationReason;omitempty"`
}

type banReq struct {
	page           web.PageReq
	server         int
	recipient      string
	startFrom      time.Time
	startTo        time.Time
	expirationFrom time.Time
	expirationTo   time.Time
}

type banModel struct {
	page           web.PageReq
	server         int
	recipient      string
	startFrom      time.Time
	startTo        time.Time
	expirationFrom time.Time
	expirationTo   time.Time
}

type banCreationReq struct {
	Server    int
	Recipient string
	Duration  time.Duration
	Reason    string
}

type banCreation struct {
	server    int
	recipient string
	giver     int
	duration  time.Duration
	reason    string
}

type banModificationReq struct {
	banCreationReq
	ModificationReason string
}

type banRemovalReq struct {
	RemovalReason string
}

const (
	OldBanExpired  = 'E'
	OldBanUnbanned = 'U'
	OldBanModified = 'M'
)
