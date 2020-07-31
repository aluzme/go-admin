package login

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/aluzme/go-admin/modules/language"
	"github.com/aluzme/go-admin/modules/logger"
)

type Login struct {
	Name string
}

func GetLoginComponent() *Login {
	return &Login{
		Name: "login",
	}
}

var DefaultFuncMap = template.FuncMap{
	"lang":     language.Get,
	"langHtml": language.GetFromHtml,
	"link": func(cdnUrl, prefixUrl, assetsUrl string) string {
		if cdnUrl == "" {
			return prefixUrl + assetsUrl
		}
		return cdnUrl + assetsUrl
	},
	"isLinkUrl": func(s string) bool {
		return (len(s) > 7 && s[:7] == "http://") || (len(s) > 8 && s[:8] == "https://")
	},
	"render": func(s, old, repl template.HTML) template.HTML {
		return template.HTML(strings.Replace(string(s), string(old), string(repl), -1))
	},
	"renderJS": func(s template.JS, old, repl template.HTML) template.JS {
		return template.JS(strings.Replace(string(s), string(old), string(repl), -1))
	},
	"divide": func(a, b int) int {
		return a / b
	},
}

func (l *Login) GetTemplate() (*template.Template, string) {
	tmpl, err := template.New("login_theme1").
		Funcs(DefaultFuncMap).
		Parse(loginTmpl)

	if err != nil {
		logger.Error("Login GetTemplate Error: ", err)
	}

	return tmpl, "login_theme1"
}

func (l *Login) GetAssetList() []string               { return AssetsList }
func (l *Login) GetAsset(name string) ([]byte, error) { return Asset(name[1:]) }
func (l *Login) IsAPage() bool                        { return true }
func (l *Login) GetName() string                      { return "login" }

func (l *Login) GetContent() template.HTML {
	buffer := new(bytes.Buffer)
	tmpl, defineName := l.GetTemplate()
	err := tmpl.ExecuteTemplate(buffer, defineName, l)
	if err != nil {
		fmt.Println("ComposeHtml Error:", err)
	}
	return template.HTML(buffer.String())
}
