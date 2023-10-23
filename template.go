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

type IndexData struct {
	LogoURL  string
	LangList []string
}

func RenderIndexPage(logoURL string) ([]byte, error) {
	tpl, err := template.New("index").Parse(string(indexHTML))
	if err != nil {
		return nil, err
	}

	result := strings.Builder{}
	data := IndexData{
		LogoURL:  logoURL,
		LangList: lexers.Names(false),
	}
	if err := tpl.Execute(&result, &data); err != nil {
		return nil, err
	}

	return []byte(result.String()), nil
}
