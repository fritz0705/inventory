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

		email, password := r.PostForm.Get("email"), r.PostForm.Get("password")

		if email == "" || password == "" {
			app.renderTemplate(w, r, map[string]interface{}{
				"email": email,
				"password": password,
			}, "Login", "Layout")
			return
		}

		var (
			dbActive bool
			dbPasswordSalt string
			dbPasswordHash string
			dbUserId string
		)
		err = app.Database.QueryRow(`SELECT 'password_hash', 'password_salt',
			'is_active', 'id' FROM 'user' WHERE 'email' = ?`, email).Scan(&dbPasswordHash,
		&dbPasswordSalt, &dbActive, &dbUserId)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		passwordHash := calculatePasswordHash(password, dbPasswordSalt)
		if passwordHash == dbPasswordHash {
			session, err := app.Sessions.Get(r, app.SessionName)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			session.Values["userId"] = dbUserId
			session.Save(r, w)
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
	switch r.Method {
	case "GET":
		app.renderTemplate(w, r, nil, "Register", "Layout")
	}
}

func calculatePasswordHash(password string, salt string) string {
	return ""
}
