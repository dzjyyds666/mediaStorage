package core

import (
	"context"
	"io"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/logx"
)

type StorageCoreServer struct {
	ctx        context.Context
	fileServer *FileIndexServer
	boxServ    *BoxServer
	depotServ  *DepotServer
}

func NewStorageCoreServer(ctx context.Context, cfg *Config, fileServer *FileIndexServer, boxServ *BoxServer, depotServ *DepotServer) *StorageCoreServer {
	return &StorageCoreServer{ctx: ctx, fileServer: fileServer, boxServ: boxServ, depotServ: depotServ}
}

// 申请文件上传
func (ss *StorageCoreServer) ApplyUpload(ctx context.Context, init *InitUpload) (string, error) {
	// 生成文件的fid
	info := init.ToMediaFileInfo()

	info.Fid = ss.fileServer.randFid()
	init.Fid = info.Fid

	return info.Fid, ss.do(
		// 检查存储文件的depot是不是创建了
		func(ctx context.Context, f *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			depotInfo, err := ss.depotServ.QueryDepotInfo(ctx, info.GetDepot().DepotId)
			if err != nil {
				logx.Errorf("StorageCoreServer|ApplyUpload|QueryDepotInfo|depotId: %s|err: %s", info.BoxInfo.GetDepotId(), err.Error())
				return err
			}
			logx.Infof("StorageCoreServer|ApplyUpload|QueryDepotInfo|depotInfo: %v", conv.ToJsonWithoutError(depotInfo))
			return nil
		},
		// 检查存储文件的box是不是创建了
		func(ctx context.Context, f *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			boxInfo, err := ss.boxServ.QueryBoxInfo(ctx, info.BoxInfo.BoxId)
			if err != nil {
				logx.Errorf("StorageCoreServer|ApplyUpload|QueryBoxInfo|boxId: %s|err: %s", info.BoxInfo.BoxId, err.Error())
				return err
			}
			logx.Infof("StorageCoreServer|ApplyUpload|QueryBoxInfo|boxInfo: %v", conv.ToJsonWithoutError(boxInfo))
			return nil
		},
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
	prepareFileInfo, err := ss.fileServer.QueryPerpareFileInfo(ctx, boxInfo.GetDepotId(), boxId, fid)
	if err != nil {
		logx.Errorf("StorageCoreServer|SingleUpload|QueryPerpareFileInfo|boxId: %s|fid: %s|err: %s", boxId, fid, err.Error())
		return err
	}

	return ss.do(
	// 开始上传文件
	)(ctx, prepareFileInfo)
}
