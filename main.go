package main

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/server"
	"github.com/maksymiliank/arrival-mc-backend/util"
	"github.com/maksymiliank/arrival-mc-backend/util/db"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"github.com/maksymiliank/arrival-mc-backend/ws"
	"log"
	"net/http"
	"os"
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
	args := os.Args[1:]
	dbFullSetup := false
	if len(args) > 0 && args[0] == "dbfullsetup" {
		dbFullSetup = true
	}

	cfg := util.ReadCfg()

	db.SetUp(cfg.DB, dbFullSetup)

	router := web.NewRouter()

	server.SetUp(router)

	wsServ := ws.SetUp(router, cfg.WS)
	authService, _ := auth.SetUp(router, wsServ)

	log.Print("The application is running!")
	log.Fatal(http.ListenAndServe(":8080", Handler{router, authService}))
}
