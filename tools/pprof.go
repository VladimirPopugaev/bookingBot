package tools

import (
	"net/http"
	_ "net/http/pprof"
)

func StartPprof(port string) error {
	return http.ListenAndServe(":"+port, nil)
}
