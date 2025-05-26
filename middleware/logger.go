package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// zap + gin: https://www.cnblogs.com/you-men/p/14694928.html#_labelTop

// GinLogger 接收gin默认的日志
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		requestURI := ctx.Request.RequestURI
		ctx.Next()
		cost := time.Since(start)
		logger.Info(
			requestURI,
			zap.Int("status", ctx.Writer.Status()),
			zap.String("method", ctx.Request.Method),
			zap.String("ip", ctx.ClientIP()),
			zap.String("user-agent", ctx.Request.UserAgent()),
			zap.String("errors", ctx.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recovery掉项目出现的panic
func GinRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if oe, ok := err.(*net.OpError); ok {
					if se, ok := oe.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(se.Error(), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				request, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(
						c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(request)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error))
					c.Abort()
					return
				}
				if stack {
					logger.Error(
						"[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(request)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error(
						"[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(request)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
