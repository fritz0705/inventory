package inventory

import (
	"database/sql"
	"net/http"
)

func (app *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		app.renderTemplate(w, r, nil, "Login", "Layout")
	case "POST":
		err := r.ParseForm()
		if err == sql.ErrNoRows {
			app.renderTemplate(w, r, nil, "Login", "Layout")
			return
		}
		if err != nil {
			app.Error(w, err)
			return
		}

		email, password := r.PostForm.Get("email"), r.PostForm.Get("password")

		if email == "" || password == "" {
			app.renderTemplate(w, r, r.PostForm, "Login", "Layout")
			return
		}

		var user User

		res, err := app.Database.Query(`SELECT * FROM 'user' WHERE 'email' = ?`, email)
		if err != nil {
			app.Error(w, err)
			return
		}
		if !res.Next() {
			// User not found
			app.renderTemplate(w, r, nil, "Login", "Layout")
			return
		}

		err = user.Load(res)
		if err != nil {
			app.Error(w, err)
			return
		}

		if user.CheckPassword(password) {
			session, err := app.Sessions.Get(r, app.SessionName)
			if err != nil {
				app.Error(w, err)
				return
			}

			session.Values["userId"] = user.Id
			session.Save(r, w)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (app *Application) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		session, err := app.Sessions.Get(r, app.SessionName)
		if err == nil && session != nil && !session.IsNew {
			delete(session.Values, "userId")
			session.Save(r, w)
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *Application) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		app.renderTemplate(w, r, nil, "Register", "Layout")
	case "POST":
		err := r.ParseForm()
		if err != nil {
			app.Error(w, err)
		}

		if r.PostForm["name"] == nil || r.PostForm["email"] == nil ||
			r.PostForm["password"] == nil || r.PostForm["password_confirmation"] == nil ||
			r.PostForm.Get("password") != r.PostForm.Get("password_confirmation") {
			app.renderTemplate(w, r, r.PostForm, "Register", "Layout")
			return
		}

		user := User{
			Name:  r.PostForm.Get("name"),
			Email: r.PostForm.Get("email"),
		}
		user.SetPassword(r.PostForm.Get("password"))

		err = user.Save(app.Database)

		if err != nil {
			app.Error(w, err)
			return
		}

		session, err := app.Sessions.Get(r, app.SessionName)
		if err != nil {
			app.Error(w, err)
			return
		}

		session.Values["userId"] = user.Id
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (app *Application) SettingsHandler(w http.ResponseWriter, r *http.Request) {
}
