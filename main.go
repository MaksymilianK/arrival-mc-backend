package main

import (
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
	dbPool := db.SetUp()

	r := web.NewRouter()
	server.SetUp(r, dbPool)

	log.Print("The application is running!")
	log.Fatal(http.ListenAndServe(":8000", Handler{r}))
}
