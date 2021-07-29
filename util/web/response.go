package web

import (
	"encoding/json"
	"fmt"
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

func ExtendSession(res http.ResponseWriter, SID string, maxAge int) {
	res.Header().Set(
		"Set-Cookie",
		fmt.Sprintf("SID=%s; max-age=%d; path=/", SID, maxAge),
	)
}

// Write tries to parse data to JSON and then write it as a response body with the OK status if no status has been set
// on response. If succeeds, callers should just return, because the response is already written. Otherwise, it logs
// a message and returns.
func Write(res http.ResponseWriter, data interface{}) {
	var bytes []byte

	var dataObj interface{}

	switch t := data.(type) {
	case string:
		dataObj = MessageRes{t}
	default:
		dataObj = data
	}

	var err error
	bytes, err = json.Marshal(dataObj)
	if err != nil {
		InternalServerError(res, err)
		return
	}
	tryWrite(res, bytes)
}


// Created writes a response with the Created status. It writes location header with a location of the created resource.
// If succeeds, callers should just return, because the response is already written. Otherwise, it logs a message
// and returns.
func Created(res http.ResponseWriter, location string) {
	res.WriteHeader(http.StatusCreated)
	res.Header().Add("Location", location)
	Write(res, "Created a new resource")
}

// MethodNotAllowed writes a response with 405 Method Not Allowed status. It writes Allow header with available methods
// given by 'allowed' argument. If succeeds, callers should just return, because the response is already written.
// Otherwise, it logs a message and returns.
func MethodNotAllowed(res http.ResponseWriter, allowed []string) {
	allowHeader(res, allowed)
	res.WriteHeader(http.StatusMethodNotAllowed)
	Write(res, "Cannot use the method")
}

// NotFound writes a response with 404 Not Found status. If succeeds, callers should just return, because the response
// is already written. Otherwise, it logs a message and returns.
func NotFound(res http.ResponseWriter) {
	res.WriteHeader(http.StatusNotFound)
	Write(res, "The resource does not exist")
}

// InternalServerError writes a response with 500 Internal Server Error status and logs a message. If succeeds, callers
// should just return, because the response is already written. Otherwise, it logs a message and returns.
func InternalServerError(res http.ResponseWriter, err error) {
	log.Println(err)
	res.WriteHeader(http.StatusInternalServerError)
	Write(res, "Cannot process the request: unexpected error occurred")
}

// BadRequest writes a response with 400 BadRequest status. If succeeds, callers should just return, because
// the response is already written. Otherwise, it logs a message and returns.
func BadRequest(res http.ResponseWriter) {
	res.WriteHeader(http.StatusBadRequest)
	Write(res, "The request is invalid")
}

// Conflict writes a response with 409 Conflict status. If succeeds, callers should just return, because the response
// is already written. Otherwise, it logs a message and returns.
func Conflict(res http.ResponseWriter) {
	res.WriteHeader(http.StatusConflict)
	Write(res, "A conflict has occurred")
}

// Unauthorized writes a response with 401 Unauthorized status. If succeeds, callers should just return, because
// the response is already written. Otherwise, it logs a message and returns.
func Unauthorized(res http.ResponseWriter) {
	res.WriteHeader(http.StatusUnauthorized)
	Write(res, "The user is not authenticated")
}

// Forbidden writes a response with 403 Forbidden status. If succeeds, callers should just return, because the response
// is already written. Otherwise, it logs a message and returns.
func Forbidden(res http.ResponseWriter) {
	res.WriteHeader(http.StatusForbidden)
	Write(res, "The user is not allowed to access the resource")
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

func OnError(res http.ResponseWriter, err error) {
	switch err {
	case ErrBadData:
		res.WriteHeader(http.StatusBadRequest)
	case ErrConflict:
		res.WriteHeader(http.StatusConflict)
	case ErrNotFound:
		res.WriteHeader(http.StatusNotFound)
	case ErrAuth:
		res.WriteHeader(http.StatusUnauthorized)
	case ErrPerm:
		res.WriteHeader(http.StatusForbidden)
	default:
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
	}
	Write(res, err.Error())
}

func tryWrite(res http.ResponseWriter, data []byte) {
	if _, err := res.Write(data); err != nil {
		log.Printf("Cannot write the response body: '%s', data: '%s'\n", err.Error(), data)
	}
}

func allowHeader(res http.ResponseWriter, allowed []string) {
	res.Header().Add("Allow", strings.Join(allowed, ", "))
}
