package middleware

import (
	"bytes"
	"net/http"
)

type responseWriterWrapper struct {
	w          http.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (i *responseWriterWrapper) Header() http.Header {
	return i.w.Header()
}

func (i *responseWriterWrapper) Write(buf []byte) (int, error) {
	i.body.Write(buf)
	return i.w.Write(buf)
}

func (i *responseWriterWrapper) WriteHeader(statusCode int) {
	i.statusCode = statusCode
	i.w.WriteHeader(statusCode)
}

func (i *responseWriterWrapper) Body() string {
	return i.body.String()
}

func (i *responseWriterWrapper) StatusCode() int {
	return i.statusCode
}
