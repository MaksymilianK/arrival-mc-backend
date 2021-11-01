package web

import (
	"encoding/json"
	"github.com/maksymiliank/arrival-mc-backend/util/validator"
	"io"
	"net/http"
	"strconv"
	"time"
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
	Page int // Current page. Starts from 0.
	Size int // Number of elements that are to be returned in a response.
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
func (p Params) Int(res http.ResponseWriter, name string) (int, error) {
	param, ok := p.Str(name)
	if !ok {
		return 0, ErrNotFound
	}

	intParam, err := strconv.Atoi(param)
	if err != nil {
		BadRequest(res)
		return 0, ErrBadData
	}
	return intParam, nil
}

// Time returns a query parameter as a time struct. If the parameter is not present or cannot be converted, returns
// false.
func (p Params) Time(res http.ResponseWriter, name string) (time.Time, error) {
	param, ok := p.Str(name)
	if !ok {
		return time.Time{}, ErrNotFound
	}

	t, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		BadRequest(res)
		return time.Time{}, ErrBadData
	}
	return time.Unix(t, 0), nil
}

// Page tries to extract pagination data from a request. If pagination is invalid, e.g. on or more of parameters
// are absent or have wrong format, writes a response with 400 Bad Request status.
func (p Params) Page(res http.ResponseWriter) (PageReq, bool) {
	page, err1 := p.Int(res, "page")
	size, err2 := p.Int(res, "size")
	if err1 == ErrNotFound || err2 == ErrNotFound {
		BadRequest(res)
		return PageReq{}, false
	} else if err1 == ErrBadData || err2 == ErrBadData {
		return PageReq{}, false
	}

	if err := validator.Validate(
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

	if err := validator.Validate(
		validator.InSlice(sortBy, allowed),
		validator.InSlice(sortOrder, SortAsc, SortDesc),
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
// if the cookie has been received; returns empty string and false if SID has not been extracted.
func ExtractSID(req *http.Request) (string, bool) {
	SID, err := req.Cookie("SID")
	if err != nil {
		return "", false
	}
	return SID.Value, true
}

// RequireSID extracts the Session ID from the request header and returns it if exists. Returns Session ID and true
// if the cookie has been received; writes 401 Unauthorized if SID has not been extracted.
func RequireSID(res http.ResponseWriter, req *http.Request) (string, bool) {
	SID, ok := ExtractSID(req)
	if !ok {
		Unauthorized(res)
		return "", false
	}
	return SID, true
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
