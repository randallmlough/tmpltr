package funcmaps

import (
	"github.com/randallmlough/tmplts"
	"net/http"
)

var RequestFuncMap = tmplts.RequestFuncMap{
	"url":    url,
	"method": method,
}

func url(r *http.Request) interface{} {
	return func() string {
		if requestIsNil(r) || r.URL == nil {
			return ""
		}
		return r.URL.Path
	}
}
func method(r *http.Request) interface{} {
	return func() string {
		if requestIsNil(r) {
			return ""
		}
		return r.Method
	}
}

func requestIsNil(r *http.Request) bool {
	if r == nil {
		return true
	}
	return false
}
