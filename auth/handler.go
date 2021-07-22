package auth

import (
	"fmt"
	web2 "github.com/maksymiliank/arrival-mc-backend/util/web"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
}

func SetUp(r *web2.Router) Service {
	crypto := NewCrypto()
	sessions := NewSessionManager(crypto)
	sessions.monitor()
	service := NewService(NewRepo(), sessions, crypto)
	handler := Handler{service}

	r.NewRoute(
		"/ranks",
		nil,
		map[string]web2.Handler{
			http.MethodGet: handler.getAll,
			http.MethodPost: handler.createOne,
		},
	)
	r.NewRoute(
		"/ranks/:id",
		[]web2.Extractor{web2.IntExtr},
		map[string]web2.Handler{
			http.MethodGet: handler.getOne,
			http.MethodDelete: handler.removeOne,
			http.MethodPut: handler.modifyOne,
		},
	)

	r.NewRoute(
		"/auth/current",
		nil,
		map[string]web2.Handler{
			http.MethodGet: handler.getCurrent,
			http.MethodDelete: handler.signOut,
			http.MethodPut: handler.signIn,
		},
	)

	return service
}

func (h Handler) getAll(res http.ResponseWriter, _ *http.Request, _ web2.PathVars) {
	web2.Write(res, h.service.allRanks())
}

func (h Handler) createOne(res http.ResponseWriter, req *http.Request, _ web2.PathVars) {
	SID, ok := web2.ExtractSID(res, req)
	if !ok {
		return
	}

	if _, err := h.service.RequirePerm(SID, ""); err != nil {
		web2.OnError(res, err)
		return
	}

	var rank rankCreation
	if !web2.Read(res, req, &rank) {
		return
	}

	ID, err := h.service.createRank(rank)
	if err != nil {
		web2.OnError(res, err)
		return
	}

	web2.Created(res, strconv.Itoa(ID))
}

func (h Handler) getOne(res http.ResponseWriter, _ *http.Request, vars web2.PathVars) {
	ID := vars["id"].(int)

	rank, err := h.service.oneRank(ID)
	if err != nil {
		web2.OnError(res, err)
		return
	}

	web2.Write(res, rank)
}

func (h Handler) modifyOne(res http.ResponseWriter, req *http.Request, vars web2.PathVars) {
	ID := vars["id"].(int)

	var rank rankModification
	if !web2.Read(res, req, &rank) {
		return
	}

	if err := h.service.modifyRank(ID, rank); err != nil {
		web2.OnError(res, err)
		return
	}

	web2.NoContent(res)
}

func (h Handler) removeOne(res http.ResponseWriter, _ *http.Request, vars web2.PathVars) {
	ID := vars["id"].(int)

	if err := h.service.removeRank(ID); err != nil {
		web2.OnError(res, err)
		return
	}

	web2.NoContent(res)
}

func (h Handler) getCurrent(res http.ResponseWriter, req *http.Request, _ web2.PathVars) {
	SID, ok := web2.ExtractSID(res, req)
	if !ok {
		return
	}

	p, err := h.service.current(SID)
	if err != nil {
		web2.OnError(res, err)
		return
	}

	web2.Write(res, p)
}

func (h Handler) signOut(res http.ResponseWriter, req *http.Request, _ web2.PathVars) {
	SID, ok := web2.ExtractSID(res, req)
	if !ok {
		web2.NoContent(res)
		return
	}

	if h.service.signOut(SID) {
		web2.Write(res, "Successfully logged out")
	} else {
		web2.NoContent(res)
	}
}

func (h Handler) signIn(res http.ResponseWriter, req *http.Request, _ web2.PathVars) {
	var data loginForm
	if !web2.Read(res, req, &data) {
		return
	}

	p, SID, err := h.service.signIn(data)
	if err != nil {
		web2.OnError(res, err)
		return
	}

	res.Header().Add("Set-Cookie", fmt.Sprintf("SID=%s;path=/", SID))
	web2.Write(res, p)
}