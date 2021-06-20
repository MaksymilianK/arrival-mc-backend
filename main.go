package main

import (
	"github.com/maksymiliank/arrival-mc-backend/web"
	"log"
	"net/http"
)

type Handler struct{}

func (Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	web.GlobalHeaders(res)

	router := web.NewRouter()
	router.Match(res, req)
}

func main() {
	log.Print("The application is running!")
	log.Fatal(http.ListenAndServe(":8000", Handler{}))
}
