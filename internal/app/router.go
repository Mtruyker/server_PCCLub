package app

import (
	"context"
	"net/http"
	"strings"
)

type routeKey struct{}

type route struct {
	method  string
	parts   []string
	handler http.HandlerFunc
}

type Router struct {
	routes []route
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	parts := strings.Fields(pattern)
	if len(parts) != 2 {
		panic("invalid route pattern: " + pattern)
	}
	method := parts[0]
	routePath := parts[1]
	r.routes = append(r.routes, route{
		method:  method,
		parts:   splitPath(routePath),
		handler: handler,
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pathParts := splitPath(req.URL.Path)
	for _, route := range r.routes {
		if req.Method != route.method {
			continue
		}
		if len(pathParts) != len(route.parts) {
			continue
		}

		vars := make(map[string]string)
		matched := true
		for i, part := range route.parts {
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				vars[part[1:len(part)-1]] = pathParts[i]
				continue
			}
			if part != pathParts[i] {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}

		req = req.WithContext(context.WithValue(req.Context(), routeKey{}, vars))
		route.handler(w, req)
		return
	}

	http.NotFound(w, req)
}

func splitPath(path string) []string {
	if path == "" {
		return []string{""}
	}
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return []string{""}
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if path == "" {
		return []string{""}
	}
	return strings.Split(path, "/")
}

func pathValue(r *http.Request, name string) string {
	vars, _ := r.Context().Value(routeKey{}).(map[string]string)
	return vars[name]
}
