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
	LogoURL    string
	IndexURL   string
	LangList   []string
	FaviconURL string
}

func renderPage(pageTemplate string, logoURL string, indexURL string, faviconURL string) ([]byte, error) {
	tpl, err := template.New("index").Parse(pageTemplate)
	if err != nil {
		return nil, err
	}

	result := strings.Builder{}
	data := IndexData{
		LogoURL:    logoURL,
		IndexURL:   indexURL,
		LangList:   lexers.Names(false),
		FaviconURL: faviconURL,
	}
	if err := tpl.Execute(&result, &data); err != nil {
		return nil, err
	}

	return []byte(result.String()), nil
}

func RenderIndexPage(logoURL string, indexURL string, faviconURL string) ([]byte, error) {
	return renderPage(string(indexHTML), logoURL, indexURL, faviconURL)
}

func RenderOutputPre(logoURL string, indexURL string, faviconURL string) ([]byte, error) {
	return renderPage(string(outputPreHTML), logoURL, indexURL, faviconURL)
}
