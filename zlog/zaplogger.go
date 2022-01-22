package zlog

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var SLogger *zap.SugaredLogger

func InitLogger() {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		fmt.Println("{\"msg\":\"FATAL: Could not get zap logger setup\",", "\"error\":\"", err, "\"}")
		panic("EXIT -1")
	}
	SLogger = Logger.Sugar()
}

// Config is config setting for Ginzap
type Config struct {
	TimeFormat string
	UTC        bool
	SkipPaths  []string
}

// Ginzap returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
//
// Requests with errors are logged using zap.Error().
// Requests without errors are logged using zap.Info().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
func Ginzap(logger *zap.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	return GinzapWithConfig(logger, &Config{TimeFormat: timeFormat, UTC: utc})
}

// GinzapWithConfig returns a gin.HandlerFunc using configs
func GinzapWithConfig(logger *zap.Logger, conf *Config) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		if _, ok := skipPaths[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)
			if conf.UTC {
				end = end.UTC()
			}

			fields := []zapcore.Field{
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Duration("latency", latency),
				zap.Int64("latency_milliseconds", latency.Milliseconds()),
				zap.Int64("latency_microseconds", latency.Microseconds()),
				zap.String("latency_string", latency.String()),
				zap.String("request_uri", c.Request.RequestURI),
				zap.Int64("content_length", c.Request.ContentLength),
				zap.String("host", c.Request.Host),
				zap.String("request_id", c.Request.Header.Get("X-Request-Id")),
			}
			if conf.TimeFormat != "" {
				fields = append(fields, zap.String("time", end.Format(conf.TimeFormat)))
			}

			for i, e := range c.Errors.Errors() {
				fields = append(fields, zap.String("error-"+strconv.Itoa(i), e))
			}

			if len(c.Errors) > 0 {
				logger.Error(path, fields...)
			} else {
				logger.Info(path, fields...)
			}
		}
	}
}

// RecoveryWithZap returns a gin.HandlerFunc (middleware)
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func RecoveryWithZap(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}
				if reflect.TypeOf(err) == reflect.TypeOf(&ErrorAPIResponse{}) {
					errResp := err.(*ErrorAPIResponse)
					if errResp.RecoveryLog {
						if stack {
							logger.WithOptions(zap.AddCallerSkip(2)).Error("[Recovery from panic]",
								zap.Time("time", time.Now().UTC()),
								zap.Any("error", err),
								zap.String("request", string(httpRequest)),
								zap.String("stack", string(debug.Stack())),
							)
						} else {
							logger.WithOptions(zap.AddCallerSkip(2)).Error("[Recovery from panic]",
								zap.Time("time", time.Now().UTC()),
								zap.Any("error", err),
								zap.String("request", string(httpRequest)),
							)
						}
					}
					c.AbortWithStatusJSON(errResp.StatusCode, err)
				} else {
					if stack {
						logger.WithOptions(zap.AddCallerSkip(2)).Error("[Recovery from panic]",
							zap.Time("time", time.Now().UTC()),
							zap.Any("error", err),
							zap.String("request", string(httpRequest)),
							zap.String("stack", string(debug.Stack())),
						)
					} else {
						logger.WithOptions(zap.AddCallerSkip(2)).Error("[Recovery from panic]",
							zap.Time("time", time.Now().UTC()),
							zap.Any("error", err),
							zap.String("request", string(httpRequest)),
						)
					}
					c.AbortWithStatusJSON(http.StatusInternalServerError, err)
				}

			}
		}()
		c.Next()
	}
}

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := c.Request.Header.Get("X-Request-Id")
		if reqId == "" {
			reqId = uuid.NewV4().String()
		}
		c.Request.Header.Set("X-Request-Id", reqId)
		c.Writer.Header().Set("X-Request-Id", reqId)
		c.Next()
	}
}
