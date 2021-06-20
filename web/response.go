package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// MessageRes is a text-only simple json response.
type MessageRes struct {
	Message string `json:"message"`
}

// PageRes contains paginated data and total amount of the requested resources in the database.
type PageRes struct {
	Total int         `json:"total"` //The number of all the elements available in the database.
	Data  interface{} `json:"data"`  //Requested data.
}

// GlobalHeaders writes http headers shared among all the server responses.
func GlobalHeaders(res http.ResponseWriter) {
	res.Header().Add("Content-Type", "application/json; charset=UTF-8")
}

// Write tries to parse data to JSON and then write it as a response body with the OK status if no status has been set
// on response. If succeeds, callers should just return, because the response is already written. Otherwise, it logs
// a message and returns.
func Write(data interface{}, res http.ResponseWriter) {
	var bytes []byte

	switch t := data.(type) {
	case string:
		bytes = []byte(t)
	default:
		var err error
		bytes, err = json.Marshal(data)
		if err != nil {
			InternalServerError(res, "Cannot parse a response body: "+err.Error())
			return
		}
	}
	tryWrite(bytes, res)
}


// Created writes a response with the Created status. It writes location header with a location of the created resource.
// If succeeds, callers should just return, because the response is already written. Otherwise, it logs a message
// and returns.
func Created(res http.ResponseWriter, location string) {
	res.WriteHeader(http.StatusCreated)
	res.Header().Add("Location", location)
	Write(MessageRes{"Created a new resource"}, res)
}

// MethodNotAllowed writes a response with 405 Method Not Allowed status. It writes Allow header with available methods
// given by 'allowed' argument. If succeeds, callers should just return, because the response is already written.
// Otherwise, it logs a message and returns.
func MethodNotAllowed(res http.ResponseWriter, allowed []string) {
	allowHeader(res, allowed)
	res.WriteHeader(http.StatusMethodNotAllowed)
	Write("Cannot use the method", res)
}

// NotFound writes a response with 404 Not Found status. If succeeds, callers should just return, because the response
// is already written. Otherwise, it logs a message and returns.
func NotFound(res http.ResponseWriter) {
	res.WriteHeader(http.StatusNotFound)
	Write("The resource does not exist", res)
}

// InternalServerError writes a response with 500 Internal Server Error status and logs a message. If succeeds, callers
// should just return, because the response is already written. Otherwise, it logs a message and returns.
func InternalServerError(res http.ResponseWriter, err string) {
	log.Println(err)
	res.WriteHeader(http.StatusInternalServerError)
	Write("Cannot process the request due to a server error", res)
}

// BadRequest writes a response with 400 BadRequest status. If succeeds, callers should just return, because
// the response is already written. Otherwise, it logs a message and returns.
func BadRequest(res http.ResponseWriter) {
	res.WriteHeader(http.StatusBadRequest)
	Write("The request is invalid", res)
}

// Conflict writes a response with 409 Conflict status. If succeeds, callers should just return, because the response
// is already written. Otherwise, it logs a message and returns.
func Conflict(res http.ResponseWriter) {
	res.WriteHeader(http.StatusConflict)
	Write("A conflict has occurred", res)
}

// Unauthorized writes a response with 401 Unauthorized status. If succeeds, callers should just return, because
// the response is already written. Otherwise, it logs a message and returns.
func Unauthorized(res http.ResponseWriter) {
	res.WriteHeader(http.StatusUnauthorized)
	Write("The user is not authenticated", res)
}

// Forbidden writes a response with 403 Forbidden status. If succeeds, callers should just return, because the response
// is already written. Otherwise, it logs a message and returns.
func Forbidden(res http.ResponseWriter) {
	res.WriteHeader(http.StatusForbidden)
	Write("The user is not allowed to access the resource", res)
}

// NoContent writes a response with 204 No Content status. If succeeds, callers should just return, because the response
// is already written. Otherwise, it logs a message and returns.
func NoContent(res http.ResponseWriter) {
	res.WriteHeader(http.StatusNoContent)
}

// AllowedNoContent writes a response with 204 No Content status and allow header with allowed methods provided
// as the argument. If succeeds, callers should just return, because the response is already written. Otherwise, it logs
// a message and returns.
func AllowedNoContent(res http.ResponseWriter, allowed []string) {
	allowHeader(res, allowed)
	res.WriteHeader(http.StatusNoContent)
}

func tryWrite(data []byte, res http.ResponseWriter) {
	if _, err := res.Write(data); err != nil {
		log.Printf("Cannot write the response body: '%s', data: '%s'\n", err.Error(), data)
	}
}

func allowHeader(res http.ResponseWriter, allowed []string) {
	res.Header().Add("Allow", strings.Join(allowed, ", "))
}
