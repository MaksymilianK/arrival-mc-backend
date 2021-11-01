package auth

import (
	"encoding/json"
	"github.com/maksymiliank/arrival-mc-backend/ws"
	"log"
)

type WSHandler struct {
	service Service
}

const (
	outboundRankListMsgType = 1000

	inboundRankListMsgType = 1000
)

func setUpWS(s *ws.Server, service Service) {
	h := WSHandler{service}

	s.AddMCHandler(inboundRankListMsgType, h.onRankList)
}

func (h WSHandler) onRankList(msg []byte) error {
	var r rankList
	if err := json.Unmarshal(msg, &r); err != nil {
		return err
	}

	f, err := h.service.GetRanks()
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(f)

	return nil
}
