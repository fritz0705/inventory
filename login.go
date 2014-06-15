package inventory

import (
	"net/http"
)

func (app *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		app.Templates.ExecuteTemplate(w, "LoginForm", nil)
	case "POST":
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func (app *Application) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := app.Sessions.Get(r, app.SessionName)
	if err == nil && session != nil && !session.IsNew{
		delete(session.Values, "userId")
		session.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (app *Application) RegisterHandler(w http.ResponseWriter, r *http.Request) {
}
