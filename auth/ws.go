package auth

import (
	"github.com/maksymiliank/arrival-mc-backend/ws"
)

type WSHandler struct {
	ws *ws.Server
	service Service
}

const (
	outboundRankListMsgType = 1000

	inboundRankListMsgType = 1000
)

func (h WSHandler) onRankList(DTO interface{}) error {
	rankList := DTO.(*rankList)

	ranks, err := h.service.ranksWithPerms(rankList.Server)
	if err != nil {
		return err
	}

	if err := h.ws.SendToMC(ranks); err != nil {
		return err
	}
	return nil
}
