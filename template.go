package makaroni

import (
	"github.com/alecthomas/chroma/lexers"
	"strings"
)

import (
	_ "embed"
	"html/template"
)

//go:embed resources/index.gohtml
var indexHTML []byte

//go:embed resources/pre.gohtml
var outputPreHTML []byte

type IndexData struct {
	LogoURL  string
	domain string
	LangList []string
}

func renderPage(pageTemplate string, logoURL string, domain string) ([]byte, error) {
	tpl, err := template.New("index").Parse(pageTemplate)
	if err != nil {
		return nil, err
	}

	result := strings.Builder{}
	data := IndexData{
		LogoURL:  logoURL,
		domain: domain,
		LangList: lexers.Names(false),
	}
	if err := tpl.Execute(&result, &data); err != nil {
		return nil, err
	}

	return []byte(result.String()), nil
}

func RenderIndexPage(logoURL string, domain string) ([]byte, error) {
	return renderPage(string(indexHTML), logoURL, domain)
}

func RenderOutputPre(logoURL string, domain string) ([]byte, error) {
	return renderPage(string(outputPreHTML), logoURL, domain)
}
