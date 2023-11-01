package main

import (
	"flag"
	"github.com/kaero/makaroni"
	"github.com/kaero/makaroni/helpers"
	"github.com/kaero/makaroni/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	address := flag.String("address", helpers.Getenv("MKRN_ADDRESS", "localhost:8080"), "Address to serve")
	multipartMaxMemoryEnv, err := strconv.ParseInt(helpers.Getenv("MKRN_MULTIPART_MAX_MEMORY", "100"), 0, 64)
	if err != nil {
		log.Fatalln(err)
	}
	multipartMaxMemory := flag.Int64("multipart-max-memory", multipartMaxMemoryEnv, "Maximum memory for multipart form parser")
	resultURLPrefix := flag.String("result-url-prefix", os.Getenv("MKRN_RESULT_URL_PREFIX"), "Upload result URL prefix.")
	logoURL := flag.String("logo-url", os.Getenv("MKRN_LOGO_URL"), "Your logo URL for the form page")
	style := flag.String("style", os.Getenv("MKRN_STYLE"), "Formatting style")
	s3Endpoint := flag.String("s3-endpoint", os.Getenv("MKRN_S3_ENDPOINT"), "S3 endpoint")
	s3Region := flag.String("s3-region", os.Getenv("MKRN_S3_REGION"), "S3 region")
	s3Bucket := flag.String("s3-bucket", os.Getenv("MKRN_S3_BUCKET"), "S3 bucket")
	s3KeyID := flag.String("s3-key-id", os.Getenv("MKRN_S3_KEY_ID"), "S3 key ID")
	s3SecretKey := flag.String("s3-secret-key", os.Getenv("MKRN_S3_SECRET_KEY"), "S3 secret key")
	debug := flag.Bool("debug", false, "Debug mode, using local file system instead of S3")
	help := flag.Bool("help", false, "Print usage")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	indexHTML, err := makaroni.RenderIndexPage(*logoURL)
	if err != nil {
		log.Fatalln(err)
	}

	var uploadFunc makaroni.UploadFunc

	if *debug {
		uploadFunc, err = makaroni.NewLocalUploader()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		uploadFunc, err = makaroni.NewUploader(*s3Endpoint, *s3Region, *s3Bucket, *s3KeyID, *s3SecretKey)
		if err != nil {
			log.Fatalln(err)
		}
	}
	mux := http.NewServeMux()
	mux.Handle("/",
		middleware.HttpNew(
			registry, nil).WrapHandler("/", &makaroni.PasteHandler{
			IndexHTML:          indexHTML,
			Upload:             uploadFunc,
			ResultURLPrefix:    *resultURLPrefix,
			Style:              *style,
			MultipartMaxMemory: *multipartMaxMemory,
		}))
	mux.Handle(
		"/metrics",
		middleware.HttpNew(
			registry, nil).
			WrapHandler("/metrics", promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{}),
			))

	server := http.Server{
		Addr:    *address,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
