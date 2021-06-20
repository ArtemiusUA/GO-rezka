package helpers

import (
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path"
	"text/template"
)

func Render(w http.ResponseWriter, tpl string, data interface{}) error {
	templatesPath := viper.GetString("TEMPLATES_PATH")
	if templatesPath == "" {
		cwd, _ := os.Getwd()
		templatesPath = path.Join(cwd, "templates")
	}
	templatesNames := []string{path.Join(templatesPath, "base.gohtml"),
		path.Join(templatesPath, "styles.gohtml")}
	templatesNames = append(templatesNames, path.Join(templatesPath, tpl))
	templates := template.Must(template.ParseFiles(templatesNames...))
	return templates.ExecuteTemplate(w, tpl, data)
}
