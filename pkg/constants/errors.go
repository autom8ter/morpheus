package constants

import (
	"github.com/palantir/stacktrace"
	"net/http"
)

var (
	ErrNotFound     = stacktrace.NewErrorWithCode(http.StatusNotFound, "not found")
	ErrUnauthorized = stacktrace.NewErrorWithCode(http.StatusUnauthorized, "unauthorized")
	ErrForbidden    = stacktrace.NewErrorWithCode(http.StatusForbidden, "forbidden")
	ErrServerError  = stacktrace.NewErrorWithCode(http.StatusInternalServerError, "internal server error")
)
