package core

import (
	"context"
	"net/url"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// 箱子的结构
type Box struct {
	BoxId      string     `json:"box_id"`
	BoxName    *string    `json:"box_name,omitempty"`
	FileNumber *int64     `json:"file_number,omitempty"`
	SpaceUsed  *int64     `json:"space_used,omitempty"`
	MetaData   url.Values `json:"meta_data,omitempty"`
}

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
