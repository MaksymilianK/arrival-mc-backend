package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeHandler(body string) Handler {
	return func(res http.ResponseWriter, req *http.Request, _ map[string]interface{}) {
		Write(body, res)
	}
}

func TestRouter(t *testing.T) {
	type RouteTest struct {
		path string
		method string
		status int
		body string
	}

	tests := []RouteTest{
		{"/route1/seg1/seg2", http.MethodGet, http.StatusOK, "0"},
		{"/route1/seg1/seg2", http.MethodPatch, http.StatusMethodNotAllowed, ""},
		{"/route1/seg1/seg2", http.MethodDelete, http.StatusOK, "1"},
		{"/route2/5/Thonem", http.MethodPost, http.StatusOK, "2"},
		{"/route2/0/Thonem", http.MethodPost, http.StatusNotFound, ""},
		{"/route2/33000/Thonem", http.MethodPost, http.StatusOK, "3"},
	}

	r := NewRouter()
	r.NewRoute("/route1/seg1/seg2", nil, map[string]Handler{
		http.MethodGet: fakeHandler("0"),
		http.MethodDelete: fakeHandler("1"),
	})
	r.NewRoute("/route2/:small-id/:name", []Extractor{SmallIDExtr, StringExtr}, map[string]Handler{
		http.MethodPost: fakeHandler("2"),
	})
	r.NewRoute("/route2/:id/Thonem", []Extractor{IDExtr}, map[string]Handler{
		http.MethodPost: fakeHandler("3"),
	})

	for i, test := range tests {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(test.method, test.path, nil)
		r.Match(res, req)

		if res.Code != test.status {
			t.Errorf("%d: expected '%d' status code, got '%d'", i, test.status, res.Code)
		}

		body := res.Body.String()
		if test.body != "" && body != test.body {
			t.Errorf("%d: expected '%s' response body, got '%s'", i, test.body, body)
		}
	}
}
