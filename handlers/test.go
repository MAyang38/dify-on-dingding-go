package handlers

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type testHandlers struct{}

var TestTandlers testHandlers

// HelloHandler 处理 /hello 路由
func (h *testHandlers) HelloHandler(ctx context.Context, c *app.RequestContext) {
	c.String(consts.StatusOK, "Hello Hertz!")
}
