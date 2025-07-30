package core

import (
	"context"
	"fmt"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var DepotPermissions = struct {
	Public     string // 公开
	PublicRead string // 公开读
	Private    string // 私有，需要进行校验
}{
	Public:     "public",
	PublicRead: "public_read",
	Private:    "private",
}

// 仓库的信息
func buildDepotInfoKey(id string) string {
	return fmt.Sprintf("media:depot:%s:info", id)
}

type Depot struct {
	DepotId        string  `json:"depot_id"`
	DepotName      *string `json:"depot_name,omitempty"`
	Permission     *string `json:"permission,omitempty"`
	PermissionHook *string `json:"permission_hook,omitempty"` // 权限钩子,是一个url类型的，可以是webhook，也可以是redis
}

// 切片服务，文件存储分为两部分 桶 => 仓库 => 箱子 => file
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

func (ds *DepotServer) CreateDepot(ctx context.Context, depot *Depot) error {
	return nil
}
