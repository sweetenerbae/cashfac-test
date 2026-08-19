package http

import (
	"fmt"
	standardhttp "net/http"
	"sync/atomic"
	"time"

	"cashfac-test/internal/platform/logger"
)

var requestSequence atomic.Uint64

type responseRecorder struct {
	standardhttp.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(standardhttp.StatusOK)
	}

	written, err := r.ResponseWriter.Write(body)
	r.bytes += written
	return written, err
}

func (r *responseRecorder) Unwrap() standardhttp.ResponseWriter {
	return r.ResponseWriter
}

func WithRequestLogging(next standardhttp.Handler) standardhttp.Handler {
	return standardhttp.HandlerFunc(func(w standardhttp.ResponseWriter, request *standardhttp.Request) {
		startedAt := time.Now()
		requestID := fmt.Sprintf("req-%06d", requestSequence.Add(1))
		w.Header().Set("X-Request-ID", requestID)
		recorder := &responseRecorder{ResponseWriter: w}

		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("PANIC", "unhandled request panic",
					logger.F("request_id", requestID),
					logger.F("error", recovered),
				)
				if recorder.status == 0 {
					standardhttp.Error(recorder, "internal server error", standardhttp.StatusInternalServerError)
				}
			}

			status := recorder.status
			if status == 0 {
				status = standardhttp.StatusOK
			}
			fields := []logger.Field{
				logger.F("request_id", requestID),
				logger.F("method", request.Method),
				logger.F("path", request.URL.RequestURI()),
				logger.F("status", status),
				logger.F("bytes", recorder.bytes),
				logger.F("duration", logger.Duration(time.Since(startedAt))),
			}

			switch {
			case status >= standardhttp.StatusInternalServerError:
				logger.Error("HTTP", "request failed", fields...)
			case status >= standardhttp.StatusBadRequest:
				logger.Warn("HTTP", "request rejected", fields...)
			default:
				logger.Info("HTTP", "request completed", fields...)
			}
		}()

		next.ServeHTTP(recorder, request)
	})
}
