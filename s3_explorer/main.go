package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	logger "github.com/sirupsen/logrus"
)

const (
	DEFAULT_GRACE_PERIOD = 5
	FILES_PER_REQUEST    = 1000
)

// Simple prometheus metrics
var (
	totalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_total_received_requests",
		Help: "The total number of received requests",
	})
	totalErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_total_errors",
		Help: "The total number of errors",
	})
)

func initLogger() {
	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logger.SetFormatter(formatter)
	logger.SetOutput(os.Stdout)
}

type Response struct {
	Files     []S3Object `json:"files"`
	NextToken *string    `json:"nextToken"`
}

func ParseNextToken(writer http.ResponseWriter, request *http.Request) (*string, error) {
	writeResponseError := func(writer http.ResponseWriter, message string) {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(message))
	}

	query := request.URL.Query()
	tokens, present := query["nextToken"]

	if !present {
		if len(query) > 0 {
			writeResponseError(writer, `{"error": "Api supports only single nextToken param"}`)
			return nil, fmt.Errorf("Api supports only single nextToken param")
		}
	}

	if present {
		if len(strings.TrimSpace(tokens[0])) == 0 {
			writeResponseError(writer, `{"error": "nextToken can't be blank"}`)
			return nil, fmt.Errorf("nextToken is blank")
		}

		if len(tokens) > 1 {
			writeResponseError(writer, `{"error": "Multiple values provided for nextToken"}`)
			return nil, fmt.Errorf("Multiple values provided for nextToken")
		}

		if len(query) > 1 {
			writeResponseError(writer, `{"error": "Api supports only single nextToken param"}`)
			return nil, fmt.Errorf("Multiple values provided for nextToken")
		}

		// Unescape next token before using it with AWS api
		token, err := url.QueryUnescape(tokens[0])
		if err != nil {
			logger.Error(err)
		}
		token = strings.ReplaceAll(token, " ", "+")
		return &token, nil
	}
	return nil, nil
}

func NewFilesHandlerFunc(client S3Client) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writeError := func() {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(`{"error": "Internal server error"}`))
		}

		logger.Infof("got request for /files endpoint from %s", request.RemoteAddr)
		totalRequests.Inc()

		nextToken, err := ParseNextToken(writer, request)
		if err != nil {
			totalErrors.Inc()
			logger.Error(err)
			dumpRequest(request)
			return
		}

		files, err, nextToken := client.GetFiles(nextToken)
		if err != nil {
			logger.Errorf("error occured during s3 api call: %v", err)
			dumpRequest(request)
			totalErrors.Inc()
			writeError()
			return
		}

		// Escape special chars in nextToken, for usage nextToken in url query string
		if nextToken != nil {
			escapedToken := url.QueryEscape(*nextToken)
			nextToken = &escapedToken
		}

		bytes, err := json.Marshal(Response{Files: files, NextToken: nextToken})
		if err != nil {
			totalErrors.Inc()
			logger.Errorf("error occured json unmarshal: %v", err)
			writeError()
			return
		}

		_, err = writer.Write(bytes)
		if err != nil {
			logger.Errorf("failed to write response %v", err)
		}
	}
}

func InitHttp(ctx context.Context, client S3Client) {
	var srv http.Server

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		totalRequests.Inc()
		_, err := fmt.Fprintf(writer, `{"time": "%d"}`, time.Now().Unix())
		if err != nil {
			logger.Error(err)
		}
	})

	http.HandleFunc("/isready", func(writer http.ResponseWriter, request *http.Request) {
		totalRequests.Inc()
		_, err, _ := client.GetFiles(nil)
		if err != nil {
			logger.Errorf("readiness probe failed: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(`{"error": "App is not ready"}`))
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/files", NewFilesHandlerFunc(client))

	go func() {
		logger.Infof("started endpoint at 8080 port")
		logger.Fatal(http.ListenAndServe(":8080", nil))
	}()

	<-ctx.Done()

	logger.Info("start shutdown")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), time.Duration(DEFAULT_GRACE_PERIOD)*time.Second)

	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		logger.Fatalf("Server Shutdown Failed:%+s", err)
	}
}

func main() {
	initLogger()

	bucket := flag.String("bucket", "", "S3 bucket ")

	filesPerRequest := flag.Int64("max_files", FILES_PER_REQUEST, "The number of files per request")

	flag.Parse()

	if len(strings.TrimSpace(*bucket)) == 0 {
		logger.Fatal("--bucket can't be empty")
	}
	// Create channel for syscalls
	// Start grecefull shutdown after first SIGINT/SIGTERM and stop after second
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		logger.Infof("system call:%+v", oscall)
		cancel()

		oscall = <-c
		logger.Fatalf("failed to finish gracefuly. system call:%+v", oscall)
	}()

	InitHttp(ctx, NewS3Client(*bucket, *filesPerRequest))
}

func dumpRequest(h *http.Request) {
	dump, _ := httputil.DumpRequest(h, true)
	fmt.Printf("----------------------------------REQUEST----------------------------------\n%s\n", string(dump))
}
