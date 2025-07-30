package core

import (
	"context"
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
func (ss *StorageCoreServer) ApplyUpload(ctx context.Context, init *MediaFileInfo) (string, error) {
	// 生成文件的fid
	init.Fid = ss.fileServer.randFid()

	return init.Fid, ss.do(
		// 检查存储文件的box是不是创建了
		func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {

			return nil
		},
		ss.fileServer.CreatePrepareFileInfo,
	)(ctx, init)
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
