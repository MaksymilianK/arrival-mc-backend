package server

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maksymiliank/arrival-mc-backend/web"
	"net/http"
)

// SetUp adds new routes and inits the whole package
func SetUp(r *web.Router, db *pgxpool.Pool) ServiceI {
	r.NewRoute(
		"/servers",
		nil,
		map[string]web.Handler{
			http.MethodGet: getAll,
		},
	)

	repo.setUp(db)
	return service
}

func getAll(res http.ResponseWriter, _ *http.Request, _ web.PathVars) {
	web.Write(
		res,
		serversRes{service.All()},
	)
}
