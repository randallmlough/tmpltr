package funcmaps

import (
	"net/http"
)

func RequestFuncMap(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"url":    URL(r),
		"method": Method(r),
	}
}
func URL(r *http.Request) func() string {
	return func() string {
		if requestIsNil(r) || r.URL == nil {
			return ""
		}
		return r.URL.Path
	}
}
func Method(r *http.Request) func() string {
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
