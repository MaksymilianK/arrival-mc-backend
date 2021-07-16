package main

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/db"
	"github.com/maksymiliank/arrival-mc-backend/server"
	"github.com/maksymiliank/arrival-mc-backend/web"
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
	db.SetUp()

	r := web.NewRouter()

	server.SetUp(r)
	auth.SetUp(r)

	log.Print("The application is running!")
	log.Fatal(http.ListenAndServe(":80", Handler{r}))
}
