package ban

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/server"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
	authService auth.Service
}

// SetUp adds new routes and initializes the whole package
func SetUp(r *web.Router, serverServer server.Service, authService auth.Service) Service {
	service := NewService(NewRepo(), serverServer, authService)
	handler := Handler{service, authService}

	r.NewRoute(
		"bans",
		nil,
		map[string]web.Handler{
			http.MethodGet: handler.getAll,
			http.MethodPost: handler.createOne,
		},
	)
	r.NewRoute(
		"bans/:id",
		[]web.Extractor{web.IntExtr},
		map[string]web.Handler{
			http.MethodGet: handler.getOne,
			http.MethodPut: handler.modifyOne,
			http.MethodDelete: handler.deleteOne,
		},
	)

	return service
}

func (h Handler) getAll(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	var banReq banReq
	params := web.ExtractParams(req)

	var err error

	if banReq.page, ok = params.Page(res); !ok {
		return
	}
	if banReq.server, err = params.Int(res, "server"); err == web.ErrBadData {
		return
	}
	banReq.recipient, _ = params.Str("recipient")
	if banReq.startFrom, err = params.Time(res, "start_from"); err == web.ErrBadData {
		return
	}
	if banReq.startTo, err = params.Time(res, "start_to"); err == web.ErrBadData {
		return
	}
	if banReq.expirationFrom, err = params.Time(res, "expiration_from"); err == web.ErrBadData {
		return
	}
	if banReq.expirationTo, err = params.Time(res, "expiration_to"); err == web.ErrBadData {
		return
	}

	page, err := h.service.all(SID, banReq)
	if err != nil {
		web.OnError(res, err)
		return
	}

	banModels := page.Data.([]*banMinModel)
	banRes := make([]banMinRes, 0)
	for _, b := range banModels {
		banRes = append(banRes, banMinRes{
			b.id,
			b.server,
			b.recipient,
			b.start.Unix(),
			b.expiration.Unix(),
			b.oldType,
		})
	}
	page.Data = banRes

	web.Write(res, page)
}

func (h Handler) createOne(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	var ban banCreationReq
	if !web.Read(res, req, &ban) {
		return
	}

	ID, err := h.service.createOne(SID, ban)
	if err != nil {
		web.OnError(res, err)
		return
	}
	web.Created(res, strconv.Itoa(ID))
}

func (h Handler) getOne(res http.ResponseWriter, req *http.Request, vars web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	ID := vars["id"].(int)

	b, err := h.service.one(SID, ID)
	if err != nil {
		web.OnError(res, err)
		return
	}

	web.Write(res, banFullRes{
		banMinRes{b.id, b.server, b.recipient, b.start.Unix(), b.expiration.Unix(), b.oldType},
		b.actualExpiration.Unix(),
		b.giver,
		b.reason,
		b.newBan,
		b.modder,
		b.modificationReason,
	})
}

func (h Handler) modifyOne(res http.ResponseWriter, req *http.Request, vars web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	ID := vars["id"].(int)

	var ban banModificationReq
	if !web.Read(res, req, &ban) {
		return
	}

	newBanID, err := h.service.modifyOne(SID, ID, ban)
	if err != nil {
		web.OnError(res, err)
		return
	}
	web.Created(res, strconv.Itoa(newBanID))
}

func (h Handler) deleteOne(res http.ResponseWriter, req *http.Request, vars web.PathVars) {
	SID, ok := web.RequireSID(res, req)
	if !ok {
		return
	}

	ID := vars["id"].(int)

	var ban banRemovalReq
	if !web.Read(res, req, &ban) {
		return
	}

	err := h.service.deleteOne(SID, ID, ban.RemovalReason)
	if err != nil {
		web.OnError(res, err)
		return
	}
	web.NoContent(res)
}

