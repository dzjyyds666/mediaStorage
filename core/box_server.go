package core

import (
	"context"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type BoxServer struct {
	ctx      context.Context
	boxRDB   *redis.Client
	boxMongo *mongo.Client
}

func NewBoxServer(ctx context.Context, conf *Config, dsServer *ds.DatabaseServer) *BoxServer {
	boxRedis, ok := dsServer.GetRedis("box")
	if !ok {
		panic("redis [box] not found")
	}
	boxMongo, ok := dsServer.GetMongo("box")
	if !ok {
		panic("mongo [box] not found")
	}
	return &BoxServer{
		ctx:      ctx,
		boxRDB:   boxRedis,
		boxMongo: boxMongo,
	}
}
