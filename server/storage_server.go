package server

import (
	"context"
	"net/http"
	"time"

	"github.com/dzjyyds666/Allspark-go/ds"

	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/api"
	"github.com/dzjyyds666/mediaStorage/internal/config"
	"github.com/dzjyyds666/mediaStorage/internal/handler"
	"github.com/dzjyyds666/mediaStorage/internal/logic"
	"github.com/dzjyyds666/mediaStorage/locale"
	"github.com/dzjyyds666/vortex/v2"
)

type StorageServer struct {
	ctx context.Context
	v   *vortex.Vortex
}

func NewStorageServer(ctx context.Context, cfg *config.Config, dsServer *ds.DatabaseServer) *StorageServer {
	s3Server := logic.NewS3Logic(ctx, cfg)
	boxServer := logic.NewBoxLogic(ctx, cfg, dsServer)
	depotServer := logic.NewDepotLogic(ctx, cfg, dsServer, boxServer)
	fileIndexServer := logic.NewFileIndexLogic(ctx, cfg, dsServer, s3Server, boxServer, depotServer)

	hcli := &http.Client{Timeout: 30 * time.Second}
	loginHandler := handler.NewLoginHandler(ctx, cfg.Server.Jwt, cfg.Server.ConsoleJwt, cfg.Admin)
	fileHandler := handler.NewFileHandler(ctx, fileIndexServer, boxServer, hcli)
	routers := api.PrepareRouters(loginHandler, fileHandler) // 创建路由

	v := vortex.BootStrap(
		ctx,
		vortex.WithPort(ptr.ToString(cfg.Port)),
		vortex.WithRouters(routers),
		vortex.WithJwtSecretKey(cfg.Server.Jwt.Secret),
		vortex.WithConsoleSecretKey(cfg.Server.ConsoleJwt.Secret),
		vortex.WithI18n(locale.V),
	)
	return &StorageServer{
		ctx: ctx,
		v:   v,
	}
}

// 启动服务
func (s *StorageServer) Start() {
	go s.v.Start()
}

// 停止服务
func (s *StorageServer) ShutDown(ctx context.Context) error {
	return nil
}
