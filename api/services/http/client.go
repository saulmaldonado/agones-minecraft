package http

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type appRoundTripper struct {
	logger            *zap.Logger
	originalTransport http.RoundTripper
}

var client *http.Client

func Init() {
	client = New(zap.L())
}

func New(logger *zap.Logger) *http.Client {
	return &http.Client{Transport: &appRoundTripper{
		logger:            logger,
		originalTransport: http.DefaultTransport,
	},
		Timeout: time.Second * 5,
	}
}

func Client() *http.Client {
	return client
}

func (r *appRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	var reqBodyBytes []byte

	if req.Body != nil {
		reqBodyBytes, _ = io.ReadAll(req.Body)
		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes))
	}

	r.logger.Info(req.URL.Path,
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.String("query", req.URL.RawQuery),
		zap.String("user-agent", req.UserAgent()),
		zap.String("time", start.Format(time.RFC3339)),
		zap.String("body", string(reqBodyBytes)),
	)

	res, err := r.originalTransport.RoundTrip(req)

	var status int
	var resBodyBytes []byte

	if res != nil {
		status = res.StatusCode
		resBodyBytes, _ = io.ReadAll(res.Body)
		res.Body.Close()
		res.Body = ioutil.NopCloser(bytes.NewBuffer(resBodyBytes))
	}

	end := time.Now()
	r.logger.Info(req.URL.Path,
		zap.String("method", req.Method),
		zap.Int("status", status),
		zap.String("path", req.URL.Path),
		zap.String("query", req.URL.RawQuery),
		zap.String("time", end.Format(time.RFC3339)),
		zap.Duration("latency", end.Sub(start)),
		zap.String("body", string(resBodyBytes)),
	)

	return res, err
}
