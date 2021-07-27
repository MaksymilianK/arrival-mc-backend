package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type PathVars map[string]interface{}

// Handler performs action on a HTTP request. Third argument is a map of request parameters with the parameter name
// as a key and its value as a map's value.
type Handler func(w http.ResponseWriter, r *http.Request, params PathVars)

// Extractor converts a request parameter (which is a string) to specific type required and returns it as the first
// value. If conversion cannot be done, e.g. due to invalid form of the parameter, it should return false as the second
// value.
type Extractor func(param string) (interface{}, bool)

// Defines handling of a HTTP request for one URL.
type route struct {
	segments []string
	extrs    []Extractor
	handlers map[string]Handler
	allowed  []string
}

// Router is a structure containing all routes.
type Router struct {
	routes []*route
}

// IntExtr is an extractor that tries to convert a request parameter to an int.
func IntExtr(param string) (interface{}, bool) {
	if val, err := strconv.Atoi(param); err == nil {
		return val, true
	} else {
		return nil, false
	}
}

// StringExtr is an extractor that converts a request parameter string. Note that since all request parameters can be
// perceived as strings, it always succeeds.
func StringExtr(param string) (interface{}, bool) {
	return param, true
}

// IDExtr is an extractor that tries to convert a request parameter to an ID, that is a positive (zero excluded) int.
func IDExtr(param string) (interface{}, bool) {
	if val, err := strconv.Atoi(param); err == nil {
		if val > 0 {
			return val, true
		}
	}
	return nil, false
}

// SmallIDExtr is an extractor that tries to convert a request parameter to a small ID, that is a positive
// (zero excluded) 16-bit int.
func SmallIDExtr(param string) (interface{}, bool) {
	if val, err := strconv.Atoi(param); err == nil {
		if val > 0 && val < 32768 {
			return val, true
		}
	}
	return nil, false
}

// NewRouter returns a new router with an initialized but empty routes slice.
func NewRouter() *Router {
	return &Router{make([]*route, 0, 16)}
}

// NewRoute adds a new route to the router. It takes request URL, parameters extractors in order and handlers map
// containing actions for every allowed method for the URL.
func (r *Router) NewRoute(path string, extrs []Extractor, handlers map[string]Handler) {
	var extrCounter int
	if extrs == nil {
		extrCounter = 0
	} else {
		extrCounter = len(extrs)
	}

	if strings.Count(path, ":") != extrCounter {
		panic(fmt.Sprintf("Number of extractors does not match number of parameters in route '%s'", path))
	}

	route := &route{
		segments: strings.Split(path, "/"),
		extrs:    extrs,
		handlers: handlers,
		allowed:  allowedMethods(handlers),
	}
	r.routes = append(r.routes, route)
}

// Match matches a request with a defined route. If there is none, it writes 404 Not Found. If there is a matching
// router, but it does not have any handler defined for the request method, it writes 405 Method Not Allowed. If the
// method was Options, it writes 204 No Content with Allow header instead.
func (r *Router) Match(res http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")
	for _, route := range r.routes {
		if matched, params := matchRoute(segments, route); matched {
			if req.Method == http.MethodOptions {
				AllowedNoContent(res, route.allowed)
			} else if methodAllowed(req.Method, route.handlers) {
				route.handlers[req.Method](res, req, params)
			} else {
				MethodNotAllowed(res, route.allowed)
			}
			return
		}
	}
	NotFound(res)
}

func methodAllowed(method string, handlers map[string]Handler) bool {
	for m := range handlers {
		if method == m {
			return true
		}
	}
	return false
}

func allowedMethods(handlers map[string]Handler) []string {
	allowed := make([]string, 0, len(handlers))
	for m := range handlers {
		allowed = append(allowed, m)
	}
	return allowed
}

func matchRoute(reqSegments []string, route *route) (bool, PathVars) {
	if len(reqSegments) != len(route.segments) {
		return false, nil
	}

	var params PathVars
	extrIndex := 0
	for i, s := range route.segments {
		if len(s) > 0 && s[0] == ':' {
			if val, ok := route.extrs[extrIndex](reqSegments[i]); ok {
				if params == nil {
					params = make(PathVars)
				}
				params[s] = val
				extrIndex++
			} else {
				return false, nil
			}
		} else if s != reqSegments[i] {
			return false, nil
		}

		if i == len(route.segments)-1 {
			return true, params
		}
	}
	return false, nil
}
