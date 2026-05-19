package templater

import (
	"bytes"
	"text/template"
)

// Render executes a Go text/template with the given data.
func Render(tmplText string, data any) (string, error) {
	tmpl, err := template.New("").Parse(tmplText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
