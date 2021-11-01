package player

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"net/http"
)

type Handler struct {
	service Service
}

// SetUp adds new routes and initializes the whole package
func SetUp(r *web.Router, authService auth.Service) Service {
	service := NewService(NewRepo(), authService)
	handler := Handler{service}

	r.NewRoute(
		"players",
		nil,
		map[string]web.Handler{
			http.MethodGet: handler.getAll,
		},
	)
	r.NewRoute(
		"players/:nick",
		[]web.Extractor{web.StringExtr},
		map[string]web.Handler{
			http.MethodGet: handler.getOne,
		},
	)

	return service
}

func (h Handler) getAll(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	params := web.ExtractParams(req)
	page, ok := params.Page(res)
	if !ok {
		return
	}

	nick, _ := params.Str("nick")

	players, err := h.service.all(SID, page, nick)
	if err != nil {
		web.OnError(res, err)
		return
	}
	web.Write(res, players)
}

func (h Handler) getOne(res http.ResponseWriter, req *http.Request, vars web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	player, err := h.service.one(SID, vars["nick"].(string))
	if err != nil {
		web.OnError(res, err)
		return
	}
	web.Write(res, player)
}
