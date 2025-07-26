package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/dzjyyds666/mediaStorage/server"
)

func main() {
	// var confPath = flag.String("c", "./conf/media.toml", "config path")
	flag.Parse()

	ctx := context.Background()
	mediaServer := server.NewMediaServer(ctx)
	go mediaServer.Start()

	// 监听退出信号
	quit := make(chan os.Signal, 1)
	// 注册要监听的信号:SIGINT(Ctrl+C)和SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

}
