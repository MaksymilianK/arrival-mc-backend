package server

import (
	web2 "github.com/maksymiliank/arrival-mc-backend/util/web"
	"net/http"
)

type Handler struct {
	service Service
}

// SetUp adds new routes and initializes the whole package
func SetUp(r *web2.Router) Service {
	service := NewService(NewRepo())
	handler := Handler{service}

	r.NewRoute(
		"servers",
		nil,
		map[string]web2.Handler{
			http.MethodGet: handler.getAll,
		},
	)

	return service
}

func (h Handler) getAll(res http.ResponseWriter, _ *http.Request, _ web2.PathVars) {
	web2.Write(res, h.service.all())
}
