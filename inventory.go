package inventory

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
)

type Application struct {
	Templates *template.Template
	Database  *sql.DB
	Sessions  sessions.Store

	SessionName string

	*http.ServeMux
}

func NewApplication() *Application {
	app := &Application{
		SessionName: "SESSION",
		ServeMux:    http.NewServeMux(),
	}

	app.SetUpRoutes()

	app.Templates = template.Must(template.ParseGlob("templates/*.html"))

	return app
}

func (app *Application) SetUpRoutes() {
	app.HandleFunc("/login", app.LoginHandler)
	app.HandleFunc("/logout", app.LogoutHandler)
	app.HandleFunc("/register", app.RegisterHandler)
	app.HandleFunc("/index", app.IndexHandler)
	app.HandleFunc("/dashboard", app.requiresUser(http.HandlerFunc(app.DashboardHandler)))
	app.HandleFunc("/", app.RootHandler)
}

func (h *Application) requiresUser(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.Sessions != nil {
			session, err := h.Sessions.Get(r, h.SessionName)
			if err == nil && !session.IsNew {
				f.ServeHTTP(w, r)
				return
			}
		}
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

func (h *Application) RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.NotFoundHandler(w, r)
		return
	}

	if h.Sessions != nil {
		session, err := h.Sessions.Get(r, h.SessionName)
		if err == nil && !session.IsNew {
			http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
			return
		}
	}

	http.Redirect(w, r, "/index", http.StatusTemporaryRedirect)
}

func (h *Application) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.renderWithLayout(w, map[string]interface{}{
		"Title": "Index",
	}, "Index", "Layout"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Application) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.Templates.ExecuteTemplate(w, "Dashboard", nil)
}

func (h *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeMux.ServeHTTP(w, r)
}
