package main

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/conn"
	"github.com/maksymiliank/arrival-mc-backend/server"
	"github.com/maksymiliank/arrival-mc-backend/util"
	"github.com/maksymiliank/arrival-mc-backend/util/db"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"log"
	"net/http"
)

type Handler struct{
	router *web.Router
}

func (h Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	web.GlobalHeaders(res)
	h.router.Match(res, req)
}

func main() {
	cfg, err := util.ReadCfg()
	if err != nil {
		panic(err)
	}
	db.SetUp(cfg.DB)

	r := web.NewRouter()
	_ = conn.SetUp(r, cfg.GameAllowedIP)

	server.SetUp(r)
	auth.SetUp(r)

	log.Print("The application is running!")
	log.Fatal(http.ListenAndServe(":80", Handler{r}))
}
