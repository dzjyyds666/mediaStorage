package main

import (
	"context"
	"flag"

	"github.com/dzjyyds666/Allspark-go/system"
	"github.com/dzjyyds666/mediaStorage/core"
	"github.com/dzjyyds666/mediaStorage/server"
)

func main() {
	var confPath = flag.String("c", "./conf/media.toml", "config path")
	flag.Parse()

	cfg, err := core.LoadConfig(*confPath)
	if nil != err {
		panic(err)
	}

	ctx := context.Background()
	mediaServer := server.NewMediaServer(ctx, cfg)
	go mediaServer.Start()
	// 优雅推出
	system.GracefulShutdown(mediaServer.ShutDown)
}
