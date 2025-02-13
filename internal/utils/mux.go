package utils

import (
	"log"
	"net/http"
	"regexp"
	"strings"
)

type PathToMethod struct {
	Key   string
	Value map[string]http.HandlerFunc
}

type CustomMux struct {
	routes []PathToMethod
}

func NewCustomMux() *CustomMux {
	return &CustomMux{
		routes: make([]PathToMethod, 0),
	}
}

func (mux *CustomMux) HandleFunc(route string, handler http.HandlerFunc) {
	parts := strings.Split(route, " ")
	if len(parts) != 2 {
		log.Fatalf("Invalid route format: %s", route)
	}
	method := parts[0]
	path := parts[1]
	mux.routes = append(mux.routes, PathToMethod{Key: path, Value: map[string]http.HandlerFunc{method: handler}})
}

func (mux *CustomMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, v := range mux.routes {
		pattern := "^" + regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`).ReplaceAllString(v.Key, `[^/]+`) + "$"
		matched, _ := regexp.MatchString(pattern, r.URL.Path)
		if matched {
			if handler, ok := v.Value[r.Method]; ok {
				handler(w, r)
				return
			}
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}
