package main

import (
	"flag"
	"github.com/kaero/makaroni"
	"log"
	"net/http"
	"net/url"
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
	domain := flag.String("domain-url", os.Getenv("MKRN_DOMAIN_URL"), "Domain url with schema.")
	domainUrl, err := url.Parse(*domain)
	if err != nil {
		log.Fatal(err)
	}

	resultSuffix := flag.String("result-url-prefix", os.Getenv("MKRN_RESULT_SUFFIX"), "Upload result suffix.")
	domainUrl.Path = *resultSuffix

	logoURL := flag.String("logo-url", os.Getenv("MKRN_LOGO_URL"), "Your logo URL for the form page")
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

	indexHTML, err := makaroni.RenderIndexPage(*logoURL, *domain)
	if err != nil {
		log.Fatalln(err)
	}

	outputPreHTML, err := makaroni.RenderOutputPre(*logoURL, *domain)
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
		ResultURL:          domainUrl.String(),
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
