package inventory

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
)

func (a *Application) renderWithLayout(w io.Writer, params map[string]interface{}, data interface{}, templates... string) error {
	var content []byte
	for _, tpl := range templates {
		buf := new(bytes.Buffer)
		tplData := map[string]interface{}{
			"Data": data,
			"Content": template.HTML(content),
			"Template": templates[0],
		}
		for key, val := range params {
			tplData[key] = val
		}
		err := a.Templates.ExecuteTemplate(buf, tpl, tplData)
		if err != nil {
			return err
		}
		content = buf.Bytes()
	}
	_, err := w.Write(content)
	return err
}

func (a *Application) renderTemplate(w http.ResponseWriter, r *http.Request, data interface{}, templates... string) {
	params := make(map[string]interface{})

	if data == nil {
		data = make(map[string]struct{})
	}

	if a.Sessions != nil {
		session, err := a.Sessions.Get(r, a.SessionName)
		println(session.Values["userId"])
		if err == nil && session.Values["userId"] != nil {
			params["User"] = session.Values["userId"].(string)
		}
	}

	err := a.renderWithLayout(w, params, data, templates...)
	if err != nil {
		panic(err)
	}
}
