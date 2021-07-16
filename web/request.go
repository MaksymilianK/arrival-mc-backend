package web

import (
	"encoding/json"
	v "github.com/maksymiliank/arrival-mc-backend/validator"
	"io"
	"net/http"
	"strconv"
)

// Params are query parameters of a HTTP request.
type Params map[string]string

// Sort represents the way data should be sorted.
type Sort struct {
	By    string // The name of the property to sort by.
	Order string // Sorting order.
}

// PageReq represents pagination data.
type PageReq struct {
	Page int   // Current page. Starts from 0.
	Size int   // Number of elements that are to be returned in a response.
}

// Sorting orders.
const (
	SortAsc  string = "ASC"
	SortDesc string = "DESC"
)

// Str returns a query parameter as a string. If the parameter is not present, returns false.
func (p Params) Str(name string) (string, bool) {
	param, ok := p[name]
	return param, ok
}

// Int returns a query parameter as an integer. If the parameter is not present or cannot be converted to int, returns
// false.
func (p Params) Int(name string) (int, bool) {
	param, ok := p.Str(name)
	if !ok {
		return 0, false
	}

	intParam, err := strconv.Atoi(param)
	if err != nil {
		return 0, false
	}
	return intParam, true
}

// Page tries to extract pagination data from a request. If pagination is invalid, e.g. on or more of parameters
// are absent or have wrong format, writes a response with 400 Bad Request status.
func (p Params) Page(res http.ResponseWriter) (PageReq, bool) {
	page, ok1 := p.Int("page")
	size, ok2 := p.Int("size")
	if !ok1 || !ok2 {
		BadRequest(res)
		return PageReq{}, false
	}

	if err := v.Validate(
		page >= 0,
		size >= 20 && size <= 100,
		page*size <= 10000,
	); err != nil {
		BadRequest(res)
		return PageReq{}, false
	}
	return PageReq{page, size}, true
}

// Sort tries to extract sorting instructions from a request. If sorting is invalid, it writes a response with
// 400 Bad Request status.
func (p Params) Sort(res http.ResponseWriter, allowed ...string) (Sort, bool) {
	sortBy, ok1 := p.Str("sortBy")
	sortOrder, ok2 := p.Str("sortOrder")

	if !ok1 || !ok2 {
		BadRequest(res)
		return Sort{}, false
	}

	if err := v.Validate(
		v.InSlice(sortBy, allowed),
		v.InSlice(sortOrder, SortAsc, SortDesc),
	); err != nil {
		BadRequest(res)
		return Sort{}, false
	}
	return Sort{sortBy, sortOrder}, true
}

// ExtractParams returns all query parameters from the request.
func ExtractParams(req *http.Request) Params {
	params := make(Params)
	for name, param := range req.URL.Query() {
		params[name] = param[0]
	}
	return params
}

// ExtractSID extracts the Session ID from the request header and returns it if exists. Returns Session ID and true
// if the cookie has been received; writes 401 Unauthorized if SID has not been extracted.
func ExtractSID(res http.ResponseWriter, req *http.Request) (string, bool) {
	SID, err := req.Cookie("SID")
	if err == nil {
		return SID.Value, true
	} else {
		Unauthorized(res)
		return "", false
	}
}

// Read tries to read data from a req body and then parse it from JSON to the structure provided as 'v' argument. If it
// fails to parse the data, it writes a response with 400 Bad Request status. If it cannot read the date, writes
// 500 Internal Server Error. Returns true if succeeded; false otherwise.
func Read(res http.ResponseWriter, req *http.Request, v interface{}) bool {
	data := make([]byte, req.ContentLength)
	if _, err := req.Body.Read(data); err != nil && err != io.EOF {
		InternalServerError(res, err)
		return false
	}

	if err := json.Unmarshal(data, v); err != nil {
		BadRequest(res)
		return false
	}
	return true
}
