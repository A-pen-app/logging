package logging

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var requestLogExcludes map[string]struct{}

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
		start := time.Now()
		ctx.Next()
		duration := time.Since(start)
		HTTP(ctx.Request.Context(),
			ctx.Request,
			&http.Response{
				StatusCode: ctx.Writer.Status(),
			},
			duration,
		)

	}
}
