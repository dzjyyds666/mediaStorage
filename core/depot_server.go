package core

import (
	"context"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// 切片服务，文件存储分为两部分 桶 =》 仓库 =》 箱子 =》 file
type DepotServer struct {
	ctx        context.Context
	depotRDB   *redis.Client
	depotMongo *mongo.Client
	boxServ    *BoxServer
}

// 仓库
func NewDepotServer(ctx context.Context, cfg *Config, dsServer *ds.DatabaseServer, boxServer *BoxServer) *DepotServer {
	depotRedis, ok := dsServer.GetRedis("depot")
	if !ok {
		panic("redis [depot] not found")
	}
	depotMongo, ok := dsServer.GetMongo("media_storage")
	if !ok {
		panic("mongo [media_storage] not found")
	}
	return &DepotServer{
		ctx:        ctx,
		depotRDB:   depotRedis,
		depotMongo: depotMongo,
		boxServ:    boxServer,
	}
}
