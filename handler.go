package makaroni

import (
	"github.com/alecthomas/chroma/quick"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
)

var contentTypeHTML = "text/html"
var contentTypeText = "text/plain"

type PasteHandler struct {
	IndexHTML          []byte
	Upload             func(key string, content string, contentType string) error
	Style              string
	ResultURLPrefix    string
	MultipartMaxMemory int64
}

func RespondServerInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(err)
}

func (p *PasteHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		w.Header().Set("Content-Type", contentTypeHTML)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(p.IndexHTML); err != nil {
			log.Println(err)
		}
		return
	}

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := req.ParseMultipartForm(p.MultipartMaxMemory); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	content := req.Form.Get("content")
	if len(content) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	syntax := req.Form.Get("syntax")
	if len(syntax) == 0 {
		syntax = "plaintext"
	}

	builder := strings.Builder{}
	// todo: customize HTML formatter
	if err := quick.Highlight(&builder, content, syntax, "html", p.Style); err != nil {
		RespondServerInternalError(w, err)
		return
	}
	html := builder.String()

	uuidV4, err := uuid.NewRandom()
	if err != nil {
		RespondServerInternalError(w, err)
		return
	}
	keyRaw := uuidV4.String()
	keyHTML := keyRaw + ".html"

	if err := p.Upload(keyRaw, content, contentTypeText); err != nil {
		RespondServerInternalError(w, err)
		return
	}

	if err := p.Upload(keyHTML, html, contentTypeHTML); err != nil {
		RespondServerInternalError(w, err)
		return
	}

	w.Header().Set("Location", p.ResultURLPrefix+keyHTML)
	w.WriteHeader(http.StatusFound)
}
