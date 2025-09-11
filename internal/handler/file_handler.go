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
	ctx  context.Context
	hcli *http.Client
	file *logic.FileIndexLogic
	box  *logic.BoxLogic
}

func NewFileHandler(ctx context.Context, file *logic.FileIndexLogic, box *logic.BoxLogic, hcli *http.Client) *FileHandler {
	return &FileHandler{
		ctx:  ctx,
		hcli: hcli,
		file: file,
		box:  box,
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

	fileInfo, err := fh.file.QueryFileInfo(ctx.GetContext(), fid)
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

	url, err := fh.file.SignFileUrl(ctx.GetContext(), fileInfo)
	if nil != err {
		logx.Errorf("HandleFile|SignFileUrl|fid: %s|err: %v", fid, err)
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

	boxInfo, err := fh.box.QueryBoxInfo(ctx.GetContext(), ptr.ToString(init.BoxId))
	if err != nil {
		logx.Errorf("StorageCoreServer|ApplyUpload|QueryBoxInfo|boxId: %s|err: %s", ptr.ToString(init.BoxId), err.Error())
		if errors.Is(err, pkg.ErrorEnums.ErrBoxNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BoxNotExist), nil)
		}
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	// 开始申请文件信息
	fid, err := fh.file.ApplyUpload(ctx.GetContext(), &init, boxInfo)
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

	boxInfo, err := fh.box.QueryBoxInfo(ctx.GetContext(), boxId)
	if nil != err {
		logx.Errorf("HandleSingleUpload|QueryBoxInfo|boxId: %s|err: %v", boxId, err)
		if errors.Is(err, pkg.ErrorEnums.ErrBoxNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.BoxNotExist), nil)
		}
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(pkg.SubStatusCodes.InternalError), nil)
	}
	err = fh.file.SingleUpload(ctx.GetContext(), boxInfo, fid, fileOpen)
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

// 获取到仓库id
func GetDepotId(ctx *vortex.Context) string {
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

	info, err := fh.file.QueryFileInfo(ctx.GetContext(), fid)
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

// 支持文件的分片上传
func (fh *FileHandler) HandleInitUpload(ctx *vortex.Context) error {

	return nil
}
