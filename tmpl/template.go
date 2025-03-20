package tmpl

import "embed"

//go:embed all:*.html
var resources embed.FS

func GetTemplate() string {
	file, err := resources.ReadFile("template.html")
	if err != nil {
		return ""
	}
	return string(file)
}
