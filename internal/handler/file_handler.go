package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/internal/logic"
	"github.com/dzjyyds666/mediaStorage/pkg"
	"github.com/dzjyyds666/vortex/v2"
	"github.com/labstack/echo/v4"
)

type FileHandler struct {
	ctx       context.Context
	coreLogic *logic.CoreLogic
	hcli      *http.Client
}

func NewFileHandler(ctx context.Context, coreLogic *logic.CoreLogic, hcli *http.Client) *FileHandler {
	return &FileHandler{
		ctx:       ctx,
		coreLogic: coreLogic,
		hcli:      hcli,
	}
}

// 获取文件
func (fh *FileHandler) HandleFile(ctx *vortex.Context) error {
	fid := ctx.Param("fid")
	if len(fid) == 0 {
		logx.Errorf("HandleFile|fid is empty")
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), echo.Map{
			"msg": "fid not be null",
		})
	}

	fileInfo, err := fh.coreLogic.QueryFileInfo(ctx.GetContext(), fid)
	if nil != err {
		logx.Errorf("HandleFile|QueryFileInfo|fid: %s|err: %v", fid, err)
		if errors.Is(err, pkg.ErrorEnums.ErrFileNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.FileNotExist), echo.Map{
				"msg": "file not exist",
			})
		} else {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(pkg.SubStatusCodes.InternalError), echo.Map{
				"msg": "query file info error",
			})
		}
	}

	url, err := fh.coreLogic.SignGetFileUrl(ctx.GetContext(), fileInfo)
	if nil != err {
		logx.Errorf("HandleFile|SignGetFileUrl|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(pkg.SubStatusCodes.InternalError), echo.Map{
			"msg": "get file url error",
		})
	}

	resp, err := fh.hcli.Get(url)
	if nil != err {
		logx.Errorf("HandleFile|Get|url: %s|err: %v", url, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(pkg.SubStatusCodes.InternalError), echo.Map{
			"msg": "get file url error",
		})
	}
	defer resp.Body.Close()

	if fileInfo.ContentType == nil {
		return vortex.HttpStreamResponse(ctx, "application/octet-stream", resp.Body)
	} else {
		return vortex.HttpStreamResponse(ctx, ptr.ToString(fileInfo.ContentType), resp.Body)
	}
}

// 申请上传
func (fh *FileHandler) HandleApplyUpload(ctx *vortex.Context) error {
	var init logic.InitUpload
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&init); err != nil {
		logx.Errorf("HandleApplyUpload|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}

	if init.BoxId == nil {
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}

	payload := ctx.GetSessionPayload()
	if payload == nil {
		logx.Errorf("HandleApplyUpload|GetSessionPayload|err|Permission Deny")
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.PermissionDeny), nil)
	}

	init.Uploader = ptr.String(payload.Uid)

	// 开始申请文件信息
	fid, err := fh.coreLogic.ApplyUpload(ctx.GetContext(), &init)
	if err != nil {
		logx.Errorf("HandleApplyUpload|ApplyUpload|fid: %s|err: %v", fid, err)
		if errors.Is(err, pkg.ErrorEnums.ErrFileExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.FileExist), nil)
		}
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	logx.Infof("HandleApplyUpload|ApplyUpload|fid: %s|fileInfo: %s", fid, conv.ToJsonWithoutError(init))
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"init_info": init,
	})
}

// 文件直接上传
func (fh *FileHandler) HandleSingleUpload(ctx *vortex.Context) error {
	fid := ctx.Param("fid")
	boxId := ctx.QueryParam("boxId")
	file, err := ctx.FormFile("file")
	if err != nil {
		logx.Errorf("HandleSingleUpload|FormFile|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}
	fileOpen, err := file.Open()
	if err != nil {
		logx.Errorf("HandleSingleUpload|Open|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}
	defer fileOpen.Close()

	if len(boxId) == 0 {
		boxId = "default"
	}

	err = fh.coreLogic.SingleUpload(ctx.GetContext(), boxId, fid, fileOpen)
	if nil != err {
		logx.Errorf("HandleSingleUpload|SingleUpload|fid: %s|err: %v", fid, err)
		if errors.Is(err, pkg.ErrorEnums.ErrBoxNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BoxNotExist), nil)
		} else if errors.Is(err, pkg.ErrorEnums.ErrNoPrepareFileInfo) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.NoPrepareFileInfo), nil)
		} else {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
		}
	}

	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"fid": fid,
	})
}

// 创建deport
func (fh *FileHandler) HandleDeportCreate(ctx *vortex.Context) error {
	var info logic.Depot
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&info); err != nil {
		logx.Errorf("HandleDeportCreate|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}

	err := fh.coreLogic.CreateDepot(ctx.GetContext(), &info)
	if nil != err {
		logx.Errorf("HandleDeportCreate|CreateDepot|depotInfo: %s|err: %v", conv.ToJsonWithoutError(info), err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"depot_id": info.DepotId,
	})
}

// 创建box
func (fh *FileHandler) HandleBoxCreate(ctx *vortex.Context) error {
	var info logic.Box
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&info); err != nil {
		logx.Errorf("HandleBoxCreate|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}

	err := fh.coreLogic.CreateBox(ctx.GetContext(), &info)
	if nil != err {
		logx.Errorf("HandleBoxCreate|CreateBox|boxInfo: %s|err: %v", conv.ToJsonWithoutError(info), err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"box_id": info.BoxId,
	})
}

// 获取到仓库id
func (fh *FileHandler) GetDepotId(ctx *vortex.Context) string {
	id := ctx.Param("depot_id")
	if len(id) == 0 {
		id = ctx.QueryParam("depot_id")
		if len(id) == 0 {
			id = ctx.Request().Header.Get("Depot-Id")
		}
	}
	if len(id) == 0 {
		id = "default"
	}
	return id
}

// 查询文件信息
func (fh *FileHandler) HandleFileInfo(ctx *vortex.Context) error {
	fid := ctx.Param("fid")
	if len(fid) == 0 {
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BadRequest), nil)
	}

	info, err := fh.coreLogic.QueryFileInfo(ctx.GetContext(), fid)
	if err != nil {
		logx.Errorf("HandleFileInfo|QueryFileInfo|fid: %s|err: %v", fid, err)
		if errors.Is(err, pkg.ErrorEnums.ErrFileNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.FileNotExist), nil)
		}
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	logx.Infof("HandleFileInfo|QueryFileInfo|fid: %s|info: %s", fid, conv.ToJsonWithoutError(info))
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, info)
}
