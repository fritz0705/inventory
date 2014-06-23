package inventory

import (
	"bytes"
	"html/template"
	"io"
)

func (a *Application) renderWithLayout(w io.Writer, data interface{}, templates... string) error {
	var content []byte
	for _, tpl := range templates {
		buf := new(bytes.Buffer)
		a.Templates.ExecuteTemplate(buf, tpl, map[string]interface{}{
			"Data": data,
			"Content": template.HTML(content),
		})
		content = buf.Bytes()
	}
	_, err := w.Write(content)
	return err
}
