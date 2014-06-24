package inventory

import (
	"net/http"
)

func (app *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		app.renderTemplate(w, r, nil, "Login", "Layout")
	case "POST":
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		session, err := app.Sessions.Get(r, app.SessionName)
		if err == nil && session.Values["userId"] != nil {
			session.Values["userId"] = 42
			session.Save(r, w)
		}
		return

		email, password := r.PostForm.Get("email"), r.PostForm.Get("password")

		var (
			dbActive bool
			dbPasswordSalt string
			dbPasswordHash string
		)
		err = app.Database.QueryRow(`SELECT 'password_hash', 'password_salt',
			'is_active' FROM 'user' WHERE 'email' = ?`, email).Scan(&dbPasswordHash,
		&dbPasswordSalt, &dbActive)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		passwordHash := calculatePasswordHash(password, dbPasswordSalt)
		if passwordHash == dbPasswordHash {
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

func calculatePasswordHash(password string, salt string) string {
	return ""
}
