package middlewares

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"time"
)

// RequestLogger 中间件函数
func RequestLogger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		startTime := time.Now()

		// 继续处理请求
		c.Next(ctx)

		// 记录请求日志
		duration := time.Since(startTime)
		method := string(c.Method())
		path := string(c.Request.URI().PathOriginal())
		statusCode := c.Response.StatusCode()
		clientIP := c.ClientIP()

		logMessage := fmt.Sprintf("[%s] %s %s %s %d %s",
			startTime.Format(time.RFC3339),
			clientIP,
			method,
			path,
			statusCode,
			duration,
		)

		hlog.Info(logMessage)
	}
}
