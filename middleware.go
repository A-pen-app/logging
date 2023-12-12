package logging

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger provides a gin middleware to log HTTP requests
func RequestLogger(excludes []string) gin.HandlerFunc {

	requestLogExcludes := map[string]struct{}{}
	for _, s := range excludes {
		requestLogExcludes[s] = struct{}{}
	}

	return func(ctx *gin.Context) {
		// Do nothing if the request URL is on the blacklist.
		url := ctx.Request.URL.EscapedPath()
		if _, exists := requestLogExcludes[url]; exists {
			return
		}
		forwardChain := strings.Split(ctx.GetHeader("X-Forwarded-For"), ",")
		remoteIP := ""
		if len(forwardChain) > 0 && forwardChain[0] != "" {
			remoteIP = forwardChain[0]
		} else {
			remoteIP = strings.Split(ctx.Request.RemoteAddr, ":")[0]
		}
		ctx.Request.Header.Add("x-forwarded-for", remoteIP)
		ctx.Request.Header.Add("true-client-ip", remoteIP)
		start := time.Now()
		ctx.Next()
		duration := time.Since(start)
		HTTP(ctx.Request.Context(),
			ctx.Request,
			&http.Response{
				StatusCode: ctx.Writer.Status(),
			},
			ctx.FullPath(),
			duration,
		)

	}
}
