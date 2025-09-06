package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/jwtx"

	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/core"
	"github.com/dzjyyds666/mediaStorage/locale"
	"github.com/dzjyyds666/mediaStorage/proto"
	"github.com/dzjyyds666/vortex/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type StorageServer struct {
	ctx        context.Context
	v          *vortex.Vortex
	coreServer *core.StorageCoreServer
	jwtToken   *core.Jwt
	consoleJwt *core.Jwt
	admin      struct {
		Username string `toml:"username" json:"username"`
		Password string `toml:"password" json:"password"`
	}
	hcli *http.Client
}

func NewStorageServer(ctx context.Context, cfg *core.Config, dsServer *ds.DatabaseServer) *StorageServer {
	s3Server := core.NewS3Server(ctx, cfg)
	boxServer := core.NewBoxServer(ctx, cfg, dsServer)
	depotServer := core.NewDepotServer(ctx, cfg, dsServer, boxServer)
	fileIndexServer := core.NewFileIndexServer(ctx, cfg, dsServer, s3Server, boxServer, depotServer)

	server := &StorageServer{
		ctx:        ctx,
		coreServer: core.NewStorageCoreServer(ctx, cfg, fileIndexServer, boxServer, depotServer, s3Server),
		jwtToken:   cfg.Server.Jwt,
		consoleJwt: cfg.Server.ConsoleJwt,
		admin: struct {
			Username string `toml:"username" json:"username"`
			Password string `toml:"password" json:"password"`
		}{
			Username: cfg.Admin.Username,
			Password: cfg.Admin.Password,
		},
		hcli: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	routers := apiPrepareRouters(server) // 创建路由
	v := vortex.BootStrap(
		ctx,
		vortex.WithPort(ptr.ToString(cfg.Port)),
		vortex.WithRouters(routers),
		vortex.WithJwtSecretKey(cfg.Server.Jwt.Secret),
		vortex.WithConsoleSecretKey(cfg.Server.ConsoleJwt.Secret),
		vortex.WithI18n(locale.V),
	)
	server.v = v

	return server
}

// 启动服务
func (s *StorageServer) Start() {
	s.v.Start()
}

// 停止服务
func (s *StorageServer) ShutDown(ctx context.Context) error {
	return nil
}

type loginReq struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// 获取文件
func (s *StorageServer) HandleFile(ctx *vortex.Context) error {
	fid := ctx.Param("fid")
	if len(fid) == 0 {
		logx.Errorf("HandleFile|fid is empty")
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), echo.Map{
			"msg": "fid not be null",
		})
	}

	fileInfo, err := s.coreServer.QueryFileInfo(ctx.GetContext(), fid)
	if nil != err {
		logx.Errorf("HandleFile|QueryFileInfo|fid: %s|err: %v", fid, err)
		if errors.Is(err, proto.ErrorEnums.ErrFileNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.FileNotExist), echo.Map{
				"msg": "file not exist",
			})
		} else {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(proto.SubStatusCodes.InternalError), echo.Map{
				"msg": "query file info error",
			})
		}
	}

	url, err := s.coreServer.SignGetFileUrl(ctx.GetContext(), fileInfo)
	if nil != err {
		logx.Errorf("HandleFile|SignGetFileUrl|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(proto.SubStatusCodes.InternalError), echo.Map{
			"msg": "get file url error",
		})
	}

	resp, err := s.hcli.Get(url)
	if nil != err {
		logx.Errorf("HandleFile|Get|url: %s|err: %v", url, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(proto.SubStatusCodes.InternalError), echo.Map{
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
func (s *StorageServer) HandleApplyUpload(ctx *vortex.Context) error {
	var init core.InitUpload
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&init); err != nil {
		logx.Errorf("HandleApplyUpload|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}

	if init.BoxId == nil {
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}

	payload := ctx.GetSessionPayload()
	if payload == nil {
		logx.Errorf("HandleApplyUpload|GetSessionPayload|err|Permission Deny")
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.PermissionDeny), nil)
	}

	init.Uploader = ptr.String(payload.Uid)

	// 开始申请文件信息
	fid, err := s.coreServer.ApplyUpload(ctx.GetContext(), &init)
	if err != nil {
		logx.Errorf("HandleApplyUpload|ApplyUpload|fid: %s|err: %v", fid, err)
		if errors.Is(err, proto.ErrorEnums.ErrFileExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.FileExist), nil)
		}
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.InternalError), nil)
	}
	logx.Infof("HandleApplyUpload|ApplyUpload|fid: %s|fileInfo: %s", fid, conv.ToJsonWithoutError(init))
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"init_info": init,
	})
}

// 文件直接上传
func (s *StorageServer) HandleSingleUpload(ctx *vortex.Context) error {
	fid := ctx.Param("fid")
	boxId := ctx.QueryParam("boxId")
	file, err := ctx.FormFile("file")
	if err != nil {
		logx.Errorf("HandleSingleUpload|FormFile|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}
	fileOpen, err := file.Open()
	if err != nil {
		logx.Errorf("HandleSingleUpload|Open|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}
	defer fileOpen.Close()

	if len(boxId) == 0 {
		boxId = "default"
	}

	err = s.coreServer.SingleUpload(ctx.GetContext(), boxId, fid, fileOpen)
	if nil != err {
		logx.Errorf("HandleSingleUpload|SingleUpload|fid: %s|err: %v", fid, err)
		if errors.Is(err, proto.ErrorEnums.ErrBoxNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BoxNotExist), nil)
		} else if errors.Is(err, proto.ErrorEnums.ErrNoPrepareFileInfo) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.NoPrepareFileInfo), nil)
		} else {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.InternalError), nil)
		}
	}

	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"fid": fid,
	})
}

// 创建deport
func (s *StorageServer) HandleDeportCreate(ctx *vortex.Context) error {
	var info core.Depot
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&info); err != nil {
		logx.Errorf("HandleDeportCreate|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}

	err := s.coreServer.CreateDepot(ctx.GetContext(), &info)
	if nil != err {
		logx.Errorf("HandleDeportCreate|CreateDepot|depotInfo: %s|err: %v", conv.ToJsonWithoutError(info), err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(proto.SubStatusCodes.InternalError), nil)
	}
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"depot_id": info.DepotId,
	})
}

// 创建box
func (s *StorageServer) HandleBoxCreate(ctx *vortex.Context) error {
	var info core.Box
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&info); err != nil {
		logx.Errorf("HandleBoxCreate|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}

	err := s.coreServer.CreateBox(ctx.GetContext(), &info)
	if nil != err {
		logx.Errorf("HandleBoxCreate|CreateBox|boxInfo: %s|err: %v", conv.ToJsonWithoutError(info), err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(proto.SubStatusCodes.InternalError), nil)
	}
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{
		"box_id": info.BoxId,
	})
}

// 获取到仓库id
func (s *StorageServer) GetDepotId(ctx *vortex.Context) string {
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
func (s *StorageServer) HandleFileInfo(ctx *vortex.Context) error {
	fid := ctx.Param("fid")
	if len(fid) == 0 {
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}

	info, err := s.coreServer.QueryFileInfo(ctx.GetContext(), fid)
	if err != nil {
		logx.Errorf("HandleFileInfo|QueryFileInfo|fid: %s|err: %v", fid, err)
		if errors.Is(err, proto.ErrorEnums.ErrFileNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.FileNotExist), nil)
		}
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.InternalError), nil)
	}
	logx.Infof("HandleFileInfo|QueryFileInfo|fid: %s|info: %s", fid, conv.ToJsonWithoutError(info))
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, info)
}
