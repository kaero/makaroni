package main

import (
	"flag"
	"github.com/kaero/makaroni"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	address := flag.String("address", os.Getenv("MKRN_ADDRESS"), "Address to serve")
	multipartMaxMemoryEnv, err := strconv.ParseInt(os.Getenv("MKRN_MULTIPART_MAX_MEMORY"), 0, 64)
	if err != nil {
		log.Fatalln(err)
	}
	multipartMaxMemory := flag.Int64("multipart-max-memory", multipartMaxMemoryEnv, "Maximum memory for multipart form parser")
	indexURL := flag.String("index-url", os.Getenv("MKRN_INDEX_URL"), "URL to the index page")
	resultURLPrefix := flag.String("result-url-prefix", os.Getenv("MKRN_RESULT_URL_PREFIX"), "Upload result URL prefix.")
	logoURL := flag.String("logo-url", os.Getenv("MKRN_LOGO_URL"), "Your logo URL for the form page")
	faviconURL := flag.String("favicon-url", os.Getenv("MKRN_FAVICON_URL"), "Your favicon URL")
	style := flag.String("style", os.Getenv("MKRN_STYLE"), "Formatting style")
	s3Endpoint := flag.String("s3-endpoint", os.Getenv("MKRN_S3_ENDPOINT"), "S3 endpoint")
	s3Region := flag.String("s3-region", os.Getenv("MKRN_S3_REGION"), "S3 region")
	s3Bucket := flag.String("s3-bucket", os.Getenv("MKRN_S3_BUCKET"), "S3 bucket")
	s3KeyID := flag.String("s3-key-id", os.Getenv("MKRN_S3_KEY_ID"), "S3 key ID")
	s3SecretKey := flag.String("s3-secret-key", os.Getenv("MKRN_S3_SECRET_KEY"), "S3 secret key")
	help := flag.Bool("help", false, "Print usage")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	indexHTML, err := makaroni.RenderIndexPage(*logoURL, *indexURL, *faviconURL)
	if err != nil {
		log.Fatalln(err)
	}

	outputPreHTML, err := makaroni.RenderOutputPre(*logoURL, *indexURL, *faviconURL)
	if err != nil {
		log.Fatalln(err)
	}

	uploadFunc, err := makaroni.NewUploader(*s3Endpoint, *s3Region, *s3Bucket, *s3KeyID, *s3SecretKey)
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", &makaroni.PasteHandler{
		IndexHTML:          indexHTML,
		OutputHTMLPre:      outputPreHTML,
		Upload:             uploadFunc,
		ResultURLPrefix:    *resultURLPrefix,
		Style:              *style,
		MultipartMaxMemory: *multipartMaxMemory,
	})

	server := http.Server{
		Addr:    *address,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
