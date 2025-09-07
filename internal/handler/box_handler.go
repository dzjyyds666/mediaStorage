package handler

import (
	"context"
	"encoding/json"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/mediaStorage/internal/logic"
	"github.com/dzjyyds666/mediaStorage/pkg"
	"github.com/dzjyyds666/vortex/v2"
	"github.com/labstack/echo/v4"
)

type BoxHandler struct {
	ctx context.Context
	box *logic.BoxServer
}

func NewBoxHandler(ctx context.Context, box *logic.BoxServer) *BoxHandler {
	return &BoxHandler{
		ctx: ctx,
		box: box,
	}
}

// 创建box
func (bh *BoxHandler) HandleBoxCreate(ctx *vortex.Context) error {
	var info logic.Box
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&info); err != nil {
		logx.Errorf("HandleBoxCreate|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}

	err := bh.box.CreateBox(ctx.GetContext(), &info)
	if nil != err {
		logx.Errorf("HandleBoxCreate|CreateBox|boxInfo: %s|err: %v", conv.ToJsonWithoutError(info), err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"box_id": info.BoxId,
	})
}
