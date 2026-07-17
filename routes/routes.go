package routes

import (
	"net/http"
	"strings"

	"go-server/api"
	"go-server/utils"
)

func Admin(app *api.App, serveLinks bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.Health)
	mux.HandleFunc("POST /api/links", app.Links)
	mux.HandleFunc("GET /api/links", app.Links)
	mux.HandleFunc("GET /api/links/{slug}", app.Link)
	mux.HandleFunc("DELETE /api/links/{slug}", app.Link)
	mux.HandleFunc("POST /api/users", app.Users)
	mux.HandleFunc("GET /api/users/me", app.Me)
	mux.HandleFunc("PUT /api/users/me", app.Me)
	mux.HandleFunc("POST /api/sessions", app.Sessions)
	mux.HandleFunc("DELETE /api/sessions", app.Sessions)
	mux.HandleFunc("GET /", home)
	mux.HandleFunc("GET /login", page("static/login.html"))
	mux.HandleFunc("GET /signup", page("static/signup.html"))
	mux.HandleFunc("GET /dashboard", page("static/dashboard.html"))
	if serveLinks {
		mux.HandleFunc("GET /go/{slug}", app.Redirect)
	}
	staticFiles := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	mux.Handle("GET /static/", noCache(staticFiles))
	return mux
}

func Links(app *api.App) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.Health)
	mux.HandleFunc("GET /{slug}", app.Redirect)
	return mux
}

func SplitByHost(linkHost string, linkHandler, adminHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(utils.Hostname(r.Host), linkHost) {
			linkHandler.ServeHTTP(w, r)
			return
		}
		adminHandler.ServeHTTP(w, r)
	})
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, "static/index.html")
}
func page(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, file)
	}
}

func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}
