package main

import (
	"context"
	"flag"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/system"
	"github.com/dzjyyds666/mediaStorage/internal/config"
	"github.com/dzjyyds666/mediaStorage/server"
)

func main() {
	var confPath = flag.String("c", "./conf/storage.toml", "config path")
	flag.Parse()

	cfg, err := config.LoadConfig(*confPath)
	if nil != err {
		panic(err)
	}

	logx.Infof("server config:%s", conv.ToJsonWithoutError(cfg))
	ctx := context.Background()

	dsServer := ds.InitDatabaseServer(ctx, cfg.Server.DBConfig, func(dbIdxs map[string]interface{}) {
		dbIdxs["system"] = 0
		dbIdxs["user"] = 1
		dbIdxs["depot"] = 2
		dbIdxs["file"] = 3
		dbIdxs["box"] = 5
		dbIdxs["media_storage"] = "media_storage"
	})

	storageServer := server.NewStorageServer(ctx, cfg, dsServer)
	go storageServer.Start()
	// 优雅推出
	system.GracefulShutdown(storageServer.ShutDown)
}
