package auth

import (
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"github.com/maksymiliank/arrival-mc-backend/ws"
	"net/http"
)

func SetUp(r *web.Router, wsServ *ws.Server) (Service, Crypto) {
	crypto := cryptoS{}
	sessions := newSessionManager(crypto)
	service := newService(SetUpRepo(), sessions, crypto)

	handler := Handler{service}
	wsHandler := WSHandler{wsServ, service}

	setUpHTTP(r, handler)
	setUpWS(wsServ, wsHandler)

	go sessions.monitor()

	return service, crypto
}

func setUpHTTP(r *web.Router, h Handler) {
	r.NewRoute(
		"ranks",
		nil,
		map[string]web.Handler{
			http.MethodGet:  h.getAll,
			http.MethodPost: h.createOne,
		},
	)
	r.NewRoute(
		"ranks/:id",
		[]web.Extractor{web.IntExtr},
		map[string]web.Handler{
			http.MethodGet:    h.getOne,
			http.MethodDelete: h.removeOne,
			http.MethodPatch:  h.modifyOne,
		},
	)

	r.NewRoute(
		"auth/current",
		nil,
		map[string]web.Handler{
			http.MethodGet:    h.getCurrent,
			http.MethodDelete: h.signOut,
			http.MethodPut:    h.signIn,
		},
	)
}

func setUpWS(s *ws.Server, w WSHandler) {
	s.AddMCHandler(inboundRankListMsgType, ws.MsgHandler{
		func() interface{} {return &rankList{}},
		w.onRankList,
	})
}