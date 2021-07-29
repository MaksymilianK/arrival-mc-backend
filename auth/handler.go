package auth

import (
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
}

func SetUp(r *web.Router) Service {
	crypto := NewCrypto()
	sessions := NewSessionManager(crypto)
	go sessions.monitor()
	service := NewService(NewRepo(), sessions, crypto)
	handler := Handler{service}

	r.NewRoute(
		"ranks",
		nil,
		map[string]web.Handler{
			http.MethodGet: handler.getAll,
			http.MethodPost: handler.createOne,
		},
	)
	r.NewRoute(
		"ranks/:id",
		[]web.Extractor{web.IntExtr},
		map[string]web.Handler{
			http.MethodGet: handler.getOne,
			http.MethodDelete: handler.removeOne,
			http.MethodPatch: handler.modifyOne,
		},
	)

	r.NewRoute(
		"auth/current",
		nil,
		map[string]web.Handler{
			http.MethodGet: handler.getCurrent,
			http.MethodDelete: handler.signOut,
			http.MethodPut: handler.signIn,
		},
	)

	return service
}

func (h Handler) getAll(res http.ResponseWriter, _ *http.Request, _ web.PathVars) {
	web.Write(res, h.service.allRanks())
}

func (h Handler) createOne(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	if _, err := h.service.RequirePerm(SID, "rank.modify"); err != nil {
		web.OnError(res, err)
		return
	}

	var rank rankCreation
	if !web.Read(res, req, &rank) {
		return
	}

	ID, err := h.service.createRank(rank)
	if err != nil {
		web.OnError(res, err)
		return
	}

	web.Created(res, strconv.Itoa(ID))
}

func (h Handler) getOne(res http.ResponseWriter, _ *http.Request, vars web.PathVars) {
	ID := vars["id"].(int)

	rank, err := h.service.oneRank(ID)
	if err != nil {
		web.OnError(res, err)
		return
	}

	web.Write(res, rank)
}

func (h Handler) modifyOne(res http.ResponseWriter, req *http.Request, vars web.PathVars) {
	ID := vars["id"].(int)

	var rank rankModification
	if !web.Read(res, req, &rank) {
		return
	}

	if err := h.service.modifyRank(ID, rank); err != nil {
		web.OnError(res, err)
		return
	}

	web.NoContent(res)
}

func (h Handler) removeOne(res http.ResponseWriter, _ *http.Request, vars web.PathVars) {
	ID := vars["id"].(int)

	if err := h.service.removeRank(ID); err != nil {
		web.OnError(res, err)
		return
	}

	web.NoContent(res)
}

func (h Handler) getCurrent(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	p, err := h.service.current(SID)
	if err != nil {
		web.OnError(res, err)
		return
	}

	web.Write(res, p)
}

func (h Handler) signOut(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		web.NoContent(res)
		return
	}

	if h.service.signOut(SID) {
		web.ExtendSession(res, SID, 0)
		web.Write(res, "Successfully logged out")
	} else {
		web.NoContent(res)
	}
}

func (h Handler) signIn(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	var data loginForm
	if !web.Read(res, req, &data) {
		return
	}

	p, SID, err := h.service.signIn(data)
	if err != nil {
		web.OnError(res, err)
		return
	}

	web.ExtendSession(res, SID, int(SessionLifetime.Seconds()))
	web.Write(res, p)
}