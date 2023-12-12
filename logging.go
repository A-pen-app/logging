// Package logging provides some helper functions to log with context
package logging

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/blendle/zapdriver"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
)

// Static configuration variables initalized at runtime.
var logLevel = LevelDebug
var projectID string
var keyRequestID = "request_id"
var keyUserID = "user_id"
var keyError = "err"
var keyScope = "scope"
var keyRemoteIP = "remote_ip"
var keyRoute = "route"

var zlogger *zap.Logger

// Initialize initializes the logger module.
func Initialize(c *Config) error {

	var err error

	if c != nil {
		logLevel = c.Level
		projectID = c.ProjectID
		keyRequestID = c.KeyRequestID
		keyUserID = c.KeyUserID
		keyError = c.KeyError
		keyScope = c.KeyScope
	}
	if projectID == "" {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zlogger, err = config.Build(zap.AddStacktrace(zap.ErrorLevel))
	} else if c.Development {
		zlogger, err = zapdriver.NewDevelopment()
	} else {
		zlogger, err = zapdriver.NewProduction()
	}
	if err != nil {
		return err
	}
	return nil
}

// Finalize finalizes the logging module.
func Finalize() {
	// Check if client and logger are valid.
	if zlogger != nil {
		zlogger.Sync()
	}
}

// HTTP is a helper function for logging API request/response
func HTTP(ctx context.Context, req *http.Request, res *http.Response, path string, latency time.Duration) {
	requestID := trace.SpanContextFromContext(ctx).TraceID().String()
	spanID := trace.SpanContextFromContext(ctx).SpanID().String()
	payload := zapdriver.NewHTTP(req, res)
	payload.Latency = latency.String()
	fields := []zapcore.Field{
		zapdriver.HTTP(payload),
		zapdriver.Label(keyRequestID, requestID),
		zapdriver.Label(keyRemoteIP, req.Header.Get("true-client-ip")),
		zapdriver.Label(keyRoute, path),
	}
	if projectID != "" {
		fields = append(fields, zapdriver.TraceContext(requestID, spanID, true, projectID)...)
	}
	userID, ok := ctx.Value(keyUserID).(string)
	if ok {
		fields = append(fields, zapdriver.Label(keyUserID, userID))
	}

	scope, ok := ctx.Value(keyScope).(string)
	if ok {
		fields = append(fields, zapdriver.Label(keyScope, scope))
	}

	zlogger.Info("request log", fields...)
}

// Critical logs a message of critical severity.
func Critical(ctx context.Context, format string, args ...interface{}) {
	zlog(ctx, LevelCritical, format, args, nil)
}

// Error logs a message of error severity.
func Error(ctx context.Context, format string, args ...interface{}) {
	zlog(ctx, LevelError, format, args, nil)
}

// Errorw logs a message with additional context
func Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	zlog(ctx, LevelError, msg, nil, keysAndValues)
}

// Warn logs a message of warning severity.
func Warn(ctx context.Context, format string, args ...interface{}) {
	zlog(ctx, LevelWarn, format, args, nil)
}

// Info logs a message of informational severity.
func Info(ctx context.Context, format string, args ...interface{}) {
	zlog(ctx, LevelInfo, format, args, nil)
}

// Infow logs a message with additional context
func Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	zlog(ctx, LevelInfo, msg, nil, keysAndValues)
}

// Debug logs a message of debugging severity.
func Debug(ctx context.Context, format string, args ...interface{}) {
	zlog(ctx, LevelDebug, format, args, nil)
}

func parseLabels(args []interface{}) []zapcore.Field {
	if len(args) == 0 {
		return nil
	}
	fields := []zapcore.Field{}
	for i := 0; i < len(args); {
		if i == len(args)-1 {
			break
		}
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); ok {
			switch keyStr {
			case "error", keyError:
				if err, ok := val.(error); ok {
					fields = append(fields, zapdriver.Label(keyError, err.Error()))
				}
			default:
				switch v := val.(type) {
				case string:
					fields = append(fields, zapdriver.Label(keyStr, v))
				case *string:
					if v != nil {
						fields = append(fields, zapdriver.Label(keyStr, *v))
					}
				case []byte:
					fields = append(fields, zapdriver.Label(keyStr, string(v)))
				case *[]byte:
					if v != nil {
						fields = append(fields, zapdriver.Label(keyStr, string(*v)))
					}
				case int:
					fields = append(fields, zapdriver.Label(keyStr, strconv.Itoa(v)))
				case int32:
					fields = append(fields, zapdriver.Label(keyStr, strconv.Itoa(int(v))))
				case int64:
					fields = append(fields, zapdriver.Label(keyStr, strconv.Itoa(int(v))))
				default:
					fields = append(fields, zapdriver.Label(keyStr, fmt.Sprintf("%+v", v)))
				}
			}
		}
		i += 2
	}
	return fields
}

func zlog(ctx context.Context, level Level, format string, args []interface{}, keysAndValues []interface{}) {
	if level <= LevelFirst || level >= LevelLast || level > logLevel {
		return
	}
	msg := fmt.Sprintf(format, args...)
	requestID := trace.SpanContextFromContext(ctx).TraceID().String()
	spanID := trace.SpanContextFromContext(ctx).SpanID().String()

	fields := []zapcore.Field{
		zapdriver.Label(keyRequestID, requestID),
		zapdriver.SourceLocation(runtime.Caller(2)),
	}
	if projectID != "" {
		fields = append(fields, zapdriver.TraceContext(requestID, spanID, true, projectID)...)
	}

	userID, ok := ctx.Value(keyUserID).(string)
	if ok {
		fields = append(fields, zapdriver.Label(keyUserID, userID))
	}

	scope, ok := ctx.Value(keyScope).(string)
	if ok {
		fields = append(fields, zapdriver.Label(keyScope, scope))
	}

	fields = append(fields, parseLabels(keysAndValues)...)
	switch level {
	case LevelInfo:
		zlogger.Info(msg, fields...)
	case LevelError:
		zlogger.Error(msg, fields...)
	case LevelCritical:
		zlogger.Fatal(msg, fields...)
	case LevelWarn:
		zlogger.Warn(msg, fields...)
	default:
		zlogger.Debug(msg, fields...)
	}
}
