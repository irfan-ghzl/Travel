package logger

import (
	"net/http"
	"time"

	"github.com/urfave/negroni"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a zap logger with colored console output.
func NewLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	l, _ := cfg.Build()
	return l
}

// HTTPLogger wraps an http.Handler with request logging using zap via negroni middleware.
func HTTPLogger(next http.Handler) http.Handler {
	zapLogger := NewLogger()
	n := negroni.New()
	n.Use(negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, nextFn http.HandlerFunc) {
		start := time.Now()
		nextFn(rw, r)
		res := rw.(negroni.ResponseWriter)
		zapLogger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.RequestURI),
			zap.Int("status", res.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("referer", r.Referer()),
			zap.String("protocol", r.Proto),
			zap.String("host", r.Host),
		)
	}))
	n.UseHandler(next)
	return n
}
