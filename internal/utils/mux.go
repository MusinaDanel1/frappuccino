package utils

import (
	"log"
	"net/http"
	"regexp"
	"strings"
)

/*
This Go code defines a custom HTTP router (CustomMux) that maps routes to HTTP methods and handlers, similar to http.ServeMux, but with some custom behavior.

	CustomMux struct: Holds a map of routes, where each route path is associated with another map of HTTP methods and their corresponding handler functions.
	NewCustomMux(): Creates and returns a new CustomMux instance with an empty routes map.
	HandleFunc(route, handler): Registers a route with an HTTP method and a handler. The route format is expected as METHOD PATH, and it splits the string to store the method and path separately.
	ServeHTTP(w, r): The main routing function. It checks if the requested URL path matches any registered route, and if the HTTP method is allowed for that route. If so, it invokes the corresponding handler; otherwise, it returns a 404 or 405 error.
*/

type CustomMux struct {
	routes map[string]map[string]http.HandlerFunc
}

func NewCustomMux() *CustomMux {
	return &CustomMux{
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

func (mux *CustomMux) HandleFunc(route string, handler http.HandlerFunc) {
	parts := strings.Split(route, " ")
	if len(parts) != 2 {
		log.Fatalf("Invalid route format: %s", route)
	}
	method := parts[0]
	path := parts[1]
	if _, ok := mux.routes[path]; !ok {
		mux.routes[path] = make(map[string]http.HandlerFunc)
	}
	mux.routes[path][method] = handler
}

func (mux *CustomMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for routePath, methods := range mux.routes {
		pattern := "^" + regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`).ReplaceAllString(routePath, `[^/]+`) + "$"
		matched, _ := regexp.MatchString(pattern, r.URL.Path)
		if matched {
			if handler, ok := methods[r.Method]; ok {
				handler(w, r)
				return
			}
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}
