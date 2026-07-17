package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Log будет доступен всему коду как синглтон.
var Log *zap.Logger = zap.NewNop()


// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
    lvl, err := zap.ParseAtomicLevel(level)
    if err != nil {
        return err
    }
    cfg := zap.NewProductionConfig()
    cfg.Level = lvl

    zl, err := cfg.Build()
    if err != nil {
        return err
    }

    Log = zl
    return nil
}

type (
    responceData struct {
        status int
        size int
    }
    loggingResponceWriter struct {
        http.ResponseWriter
        responceData *responceData
    }
)

func(r *loggingResponceWriter) Write(b []byte) (int,error) {
    size,err := r.ResponseWriter.Write(b)
    r.responceData.size = size
    return size,err
}

func(r *loggingResponceWriter) WriteHeader(code int) {
    r.ResponseWriter.WriteHeader(code)
    r.responceData.status = code
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        data := &responceData{
            status:0,
            size:0,
        }
        resp := loggingResponceWriter{
            w,
            data,
        }

        h.ServeHTTP(&resp, r) // обслуживание оригинального запроса

        duration := time.Since(start)

        Log.Info("got incoming HTTP request",
            zap.String("uri", r.RequestURI),
            zap.String("method", r.Method),
            zap.Any("duration", duration),
        )
        Log.Info("send massage",
            zap.Int("status", data.status),
            zap.Int("size", data.size),
        )
    }
} 