package core

import "context"

type StorageCoreServer struct {
	ctx        context.Context
	fileServer *FileIndexServer
}

func NewStorageCoreServer(ctx context.Context, cfg *Config, fileServer *FileIndexServer) *StorageCoreServer {
	return &StorageCoreServer{ctx: ctx, fileServer: fileServer}
}
