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

type DepotHandler struct {
	ctx   context.Context
	depot *logic.DepotLogic
}

func NewDepotHandler(ctx context.Context, depot *logic.DepotLogic) *DepotHandler {
	return &DepotHandler{
		ctx:   ctx,
		depot: depot,
	}
}

// 创建deport
func (dh *DepotHandler) HandleDeportCreate(ctx *vortex.Context) error {
	var info logic.Depot
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&info); err != nil {
		logx.Errorf("HandleDeportCreate|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}

	depot, err := dh.depot.CreateDepot(ctx.GetContext(), &info)
	if nil != err {
		logx.Errorf("HandleDeportCreate|CreateDepot|depotInfo: %s|err: %v", conv.ToJsonWithoutError(info), err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"depot_info": depot,
	})
}
