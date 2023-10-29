package main

import (
	_ "embed"
	"flag"
	"github.com/alecthomas/chroma/quick"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed resources/index.html
var indexHTML []byte
var contentTypeHTML = "text/html"
var contentTypeText = "text/plain"

func ServeForm(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", contentTypeHTML)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(indexHTML); err != nil {
		log.Fatalln(err)
	}
}

type PastePostHandler struct {
	Uploader   *s3manager.Uploader
	BucketRaw  string
	BucketHTML string
}

func (p *PastePostHandler) Upload(bucket string, key string, content string, contentType string) error {
	_, err := p.Uploader.Upload(&s3manager.UploadInput{
		Bucket:      &bucket,
		Key:         &key,
		ContentType: &contentType,
		Body:        strings.NewReader(content),
	})
	return err
}

func RespondServerInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(err)
}

func (p *PastePostHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := req.ParseForm(); err != nil {
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
	if err := quick.Highlight(&builder, content, syntax, "html", "github"); err != nil {
		RespondServerInternalError(w, err)
		return
	}
	html := builder.String()

	uuidV4, err := uuid.NewRandom()
	if err != nil {
		RespondServerInternalError(w, err)
		return
	}
	key := uuidV4.String()

	if err := p.Upload(p.BucketRaw, key, content, contentTypeText); err != nil {
		RespondServerInternalError(w, err)
		return
	}

	if err := p.Upload(p.BucketHTML, key, html, contentTypeHTML); err != nil {
		RespondServerInternalError(w, err)
		return
	}

	w.Header().Set("Location", "/paste/html/"+key)
	w.WriteHeader(http.StatusFound)
}

func main() {
	address := flag.String("address", os.Getenv("MKRN_ADDRESS"), "Address to serve")
	s3Endpoint := flag.String("s3-endpoint", os.Getenv("MKRN_S3_ENDPOINT"), "S3 endpoint")
	s3Region := flag.String("s3-region", os.Getenv("MKRN_S3_REGION"), "S3 region")
	s3BucketRaw := flag.String("s3-bucket-raw", os.Getenv("MKRN_S3_BUCKET_RAW"), "S3 bucket to keep RAW content")
	s3BucketHTML := flag.String("s3-bucket-html", os.Getenv("MKRN_S3_BUCKET_HTML"), "S3 bucket to keep HTML content")
	s3KeyID := flag.String("s3-key-id", os.Getenv("MKRN_S3_KEY_ID"), "S3 key ID")
	s3SecretKey := flag.String("s3-secret-key", os.Getenv("MKRN_S3_SECRET_KEY"), "S3 secret key")
	help := flag.Bool("help", false, "Print usage")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	awsSession, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(*s3KeyID, *s3SecretKey, ""),
		Endpoint:    s3Endpoint,
		Region:      s3Region,
	})
	if err != nil {
		log.Fatalln(err)
	}
	uploader := s3manager.NewUploader(awsSession)

	mux := http.NewServeMux()
	mux.HandleFunc("/", ServeForm)
	mux.Handle("/paste", &PastePostHandler{
		Uploader:   uploader,
		BucketRaw:  *s3BucketRaw,
		BucketHTML: *s3BucketHTML,
	})

	server := http.Server{
		Addr:    *address,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
