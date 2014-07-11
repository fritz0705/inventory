package inventory

import (
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

var SessionKey = []byte("inventory.bin")

type Application struct {
	Templates *template.Template
	DB        *sqlx.DB
	Sessions  sessions.Store

	// Configuration options
	SessionName   string
	TemplatesPath string
	AssetsPath    string

	*http.ServeMux

	initialized bool
}

func NewApplication() *Application {
	app := &Application{
		SessionName:   "SESSION",
		AssetsPath:    "assets/",
		TemplatesPath: "templates/",
	}

	return app
}

func (app *Application) Init() {
	app.initialized = true

	app.setUpRoutes()

	app.Templates = template.New("")
	app.Templates.Funcs(templateFuncs)
	template.Must(app.Templates.ParseGlob(app.TemplatesPath + "/*.html"))

	if app.Sessions == nil {
		app.Sessions = sessions.NewCookieStore(SessionKey)
	}
}

func (app *Application) setUpRoutes() {
	app.ServeMux = http.NewServeMux()

	app.HandleFunc("/index", app.IndexHandler)
	app.HandleFunc("/dashboard", app.requiresUser(http.HandlerFunc(app.DashboardHandler)))

	app.HandleFunc("/login", app.requiresSessions(http.HandlerFunc(app.LoginHandler)))
	app.HandleFunc("/logout", app.requiresSessions(http.HandlerFunc(app.LogoutHandler)))
	app.HandleFunc("/register", app.requiresSessions(http.HandlerFunc(app.RegisterHandler)))
	app.HandleFunc("/settings", app.SettingsHandler)

	app.HandleFunc("/search", app.SearchHandler)

	app.HandleFunc("/parts", app.ListPartsHandler)
	app.HandleFunc("/parts/", app.ShowPartHandler)
	app.HandleFunc("/parts/new", app.NewPartHandler)
	app.HandleFunc("/parts/edit/", app.EditPartHandler)
	app.HandleFunc("/parts/empty/", app.EmptyPartHandler)
	app.HandleFunc("/parts/record/", app.CreatePartAmountHandler)
	app.HandleFunc("/parts/delete/", app.DeletePartHandler)

	app.HandleFunc("/categories", app.ListCategoriesHandler)
	app.HandleFunc("/categories/new", app.NewCategoryHandler)
	app.HandleFunc("/categories/edit/", app.EditCategoryHandler)
	app.HandleFunc("/categories/delete/", app.DeleteCategoryHandler)

	app.HandleFunc("/places", app.ListPlacesHandler)
	app.HandleFunc("/places/new", app.NewPlaceHandler)
	app.HandleFunc("/places/delete/", app.DeletePlaceHandler)

	app.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(app.AssetsPath))))

	app.HandleFunc("/", app.RootHandler)
}

func (h *Application) requiresUser(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.Sessions != nil {
			session, err := h.Sessions.Get(r, h.SessionName)
			if err == nil && session.Values["userId"] != nil {
				f.ServeHTTP(w, r)
				return
			}
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func (h *Application) requiresSessions(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.Sessions != nil {
			f.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
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
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/index", http.StatusFound)
}

func (h *Application) IndexHandler(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, r, nil, "Index", "Layout")
}

func (h *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.initialized {
		h.Init()
	}
	h.ServeMux.ServeHTTP(w, r)
}
