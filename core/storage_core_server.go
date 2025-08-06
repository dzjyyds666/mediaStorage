package core

import (
	"context"
	"io"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
)

type StorageCoreServer struct {
	ctx        context.Context
	fileServer *FileIndexServer
	boxServ    *BoxServer
	depotServ  *DepotServer
	s3Server   *S3Server
}

func NewStorageCoreServer(ctx context.Context, cfg *Config, fileServer *FileIndexServer, boxServ *BoxServer, depotServ *DepotServer, s3Server *S3Server) *StorageCoreServer {
	return &StorageCoreServer{ctx: ctx, fileServer: fileServer, boxServ: boxServ, depotServ: depotServ, s3Server: s3Server}
}

// 申请文件上传
func (ss *StorageCoreServer) ApplyUpload(ctx context.Context, init *InitUpload) (string, error) {
	// 生成文件的fid
	info := init.ToMediaFileInfo()
	info.Fid = randFid()
	init.Fid = info.Fid

	boxInfo, err := ss.boxServ.QueryBoxInfo(ctx, ptr.ToString(init.BoxId))
	if err != nil {
		logx.Errorf("StorageCoreServer|ApplyUpload|QueryBoxInfo|boxId: %s|err: %s", ptr.ToString(init.BoxId), err.Error())
		return "", err
	}
	info.BoxInfo = boxInfo

	return info.Fid, ss.do(
		ss.fileServer.CreatePrepareFileInfo,
		func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			logx.Infof("StorageCoreServer|ApplyUpload|CreatePrepareFileInfo|info: %v", conv.ToJsonWithoutError(info))
			return nil
		},
	)(ctx, info)
}

func (ss *StorageCoreServer) do(funcs ...FileOption) FileOption {
	return func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
		for _, f := range funcs {
			if err := f(ss.ctx, info, opts...); err != nil {
				return err
			}
		}
		return nil
	}
}

// 文件直接上传
func (ss *StorageCoreServer) SingleUpload(ctx context.Context, boxId, fid string, r io.Reader) error {
	// 查询对应的box
	boxInfo, err := ss.boxServ.QueryBoxInfo(ctx, boxId)
	if err != nil {
		logx.Errorf("StorageCoreServer|SingleUpload|QueryBoxInfo|boxId: %s|err: %s", boxId, err.Error())
		return err
	}

	// 查询文件的初始化上传信息
	prepareFileInfo, err := ss.fileServer.QueryPerpareFileInfo(ctx, ptr.ToString(boxInfo.DepotId), boxId, fid)
	if err != nil {
		logx.Errorf("StorageCoreServer|SingleUpload|QueryPerpareFileInfo|boxId: %s|fid: %s|err: %s", boxId, fid, err.Error())
		return err
	}

	return ss.do(
		// 开始上传文件
		func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			err := ss.fileServer.SaveFileData(ctx, info, r)
			if nil != err {
				logx.Errorf("StorageCoreServer|SingleUpload|SaveFileData|boxId: %s|fid: %s|err: %s", boxId, fid, err.Error())
				return err
			}
			return nil
		},
		// 完成上传之后的文件信息构建
		ss.fileServer.CompleteUpload,
	)(ctx, prepareFileInfo)
}

func (ss *StorageCoreServer) SignGetFileUrl(ctx context.Context, info *MediaFileInfo) (string, error) {
	objectKey := info.BuildObjectKey()
	presignedURL, err := ss.s3Server.GetPresignedURL(ctx, objectKey)
	if err != nil {
		logx.Errorf("StorageCoreServer|SignGetFileUrl|GetPresignedURL|fid: %s|err: %s", info.Fid, err.Error())
		return "", err
	}
	return presignedURL, nil
}

// 查询文件的信息
func (ss *StorageCoreServer) QueryFileInfo(ctx context.Context, fid string) (*MediaFileInfo, error) {
	return ss.fileServer.QueryFileInfo(ctx, fid)
}

// 创建depot
func (ss *StorageCoreServer) CreateDepot(ctx context.Context, info *Depot) error {
	if len(info.DepotId) == 0 {
		info.DepotId = "di_" + generateRandomString(8)
	}
	return ss.depotServ.CreateDepot(ctx, info)
}

// 创建box
func (ss *StorageCoreServer) CreateBox(ctx context.Context, info *Box) error {
	if len(info.BoxId) == 0 {
		info.BoxId = "bi_" + generateRandomString(8)
	}
	return ss.boxServ.CreateBox(ctx, info)
}
