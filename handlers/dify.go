package handlers

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type difyHandlers struct{}

var DifyTandlers difyHandlers

// DifyTandlers 处理 /chat-message 路由
func (h *difyHandlers) ChatMessageHandler(ctx context.Context, c *app.RequestContext) {
	c.String(consts.StatusOK, "Hello Hertz!")
}
