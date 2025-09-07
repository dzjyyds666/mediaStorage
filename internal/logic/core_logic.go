package logic

import (
	"context"
	"io"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/internal/config"
)

// 核心的处理逻辑
type CoreLogic struct {
	ctx        context.Context
	fileServer *FileIndexServer
	boxServ    *BoxServer
	depotServ  *DepotServer
	s3Server   *S3Server
}

func NewStorageCoreServer(ctx context.Context, cfg *config.Config, fileServer *FileIndexServer, boxServ *BoxServer, depotServ *DepotServer, s3Server *S3Server) *CoreLogic {
	return &CoreLogic{ctx: ctx, fileServer: fileServer, boxServ: boxServ, depotServ: depotServ, s3Server: s3Server}
}

// 申请文件上传
func (cl *CoreLogic) ApplyUpload(ctx context.Context, init *InitUpload) (string, error) {
	// 生成文件的fid
	info := init.ToMediaFileInfo()
	info.Fid = randFid()
	init.Fid = info.Fid

	boxInfo, err := cl.boxServ.QueryBoxInfo(ctx, ptr.ToString(init.BoxId))
	if err != nil {
		logx.Errorf("StorageCoreServer|ApplyUpload|QueryBoxInfo|boxId: %s|err: %s", ptr.ToString(init.BoxId), err.Error())
		return "", err
	}
	info.Box = boxInfo

	return info.Fid, cl.do(
		cl.fileServer.CreatePrepareFileInfo,
		func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			logx.Infof("StorageCoreServer|ApplyUpload|CreatePrepareFileInfo|info: %v", conv.ToJsonWithoutError(info))
			return nil
		},
	)(ctx, info)
}

func (cl *CoreLogic) do(funcs ...FileOption) FileOption {
	return func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
		for _, f := range funcs {
			if err := f(cl.ctx, info, opts...); err != nil {
				return err
			}
		}
		return nil
	}
}

// 文件直接上传
func (cl *CoreLogic) SingleUpload(ctx context.Context, boxId, fid string, r io.Reader) error {
	// 查询对应的box
	boxInfo, err := cl.boxServ.QueryBoxInfo(ctx, boxId)
	if err != nil {
		logx.Errorf("StorageCoreServer|SingleUpload|QueryBoxInfo|boxId: %s|err: %s", boxId, err.Error())
		return err
	}

	// 查询文件的初始化上传信息
	prepareFileInfo, err := cl.fileServer.QueryPerpareFileInfo(ctx, ptr.ToString(boxInfo.DepotId), boxId, fid)
	if err != nil {
		logx.Errorf("StorageCoreServer|SingleUpload|QueryPerpareFileInfo|boxId: %s|fid: %s|err: %s", boxId, fid, err.Error())
		return err
	}

	return cl.do(
		// 开始上传文件
		func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			err := cl.fileServer.SaveFileData(ctx, info, r)
			if nil != err {
				logx.Errorf("StorageCoreServer|SingleUpload|SaveFileData|boxId: %s|fid: %s|err: %s", boxId, fid, err.Error())
				return err
			}
			return nil
		},
		// 完成上传之后的文件信息构建
		cl.fileServer.CompleteUpload,
	)(ctx, prepareFileInfo)
}
