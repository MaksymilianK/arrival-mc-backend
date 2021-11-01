package main

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/ban"
	"github.com/maksymiliank/arrival-mc-backend/player"
	"github.com/maksymiliank/arrival-mc-backend/server"
	"github.com/maksymiliank/arrival-mc-backend/util"
	"github.com/maksymiliank/arrival-mc-backend/util/db"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"github.com/maksymiliank/arrival-mc-backend/ws"
	"log"
	"net/http"
)

type Handler struct {
	router      *web.Router
	authService auth.Service
}

func (h Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	web.GlobalHeaders(res)
	if SID, ok := web.ExtractSID(req); ok {
		if h.authService.TryExtendSession(SID) {
			web.ExtendSession(res, SID, int(auth.SessionLifetime.Seconds()))
		} else {
			web.ExtendSession(res, SID, 0)
		}
	}

	h.router.Match(res, req)
}

func main() {
	cfg, err := util.ReadCfg()
	if err != nil {
		panic(err)
	}
	db.SetUp(cfg.DB)

	r := web.NewRouter()

	serverService := server.SetUp(r)

	w := ws.SetUp(r, cfg.WS)

	authService := auth.SetUp(r, w)

	player.SetUp(r, authService)

	ban.SetUp(r, serverService, authService)

	log.Print("The application is running!")
	log.Fatal(http.ListenAndServe(":8080", Handler{r, authService}))
}
