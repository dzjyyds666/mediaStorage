package core

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/mediaStorage/proto"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
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

func randDepotId() string {
	return "di_" + uuid.NewString()
}

type Depot struct {
	DepotId        string     `json:"depot_id" bson:"_id"`
	DepotName      *string    `json:"depot_name,omitempty" bson:"depot_name,omitempty"`
	Permission     *string    `json:"permission,omitempty" bson:"permission,omitempty"`
	PermissionHook *string    `json:"permission_hook,omitempty" bson:"permission_hook,omitempty"` // 权限钩子,是一个url类型的，可以是webhook，也可以是redis
	MetaData       url.Values `json:"meta_data,omitempty" bson:"meta_data,omitempty"`             // 元数据
}

// 切片服务，文件存储分为两部分 桶 => 仓库 => 箱子 => file
type DepotServer struct {
	ctx        context.Context
	depotRDB   *redis.Client
	depotMongo *mongo.Database
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

// 创建仓库
func (ds *DepotServer) CreateDepot(ctx context.Context, depot *Depot) error {
	_, err := ds.depotMongo.Collection(proto.DatabaseName.DepotDataBaseName).InsertOne(ctx, depot)
	if nil != err {
		logx.Errorf("DepotServer|CreateDepot|InsertOne|err: %v", err)
		return err
	}
	logx.Infof("DepotServer|CreateDepot|success|depot_info: %s", conv.ToJsonWithoutError(depot))
	return nil
}

// 查询仓库信息
func (ds *DepotServer) QueryDepotInfo(ctx context.Context, depotId string) (*Depot, error) {
	if depotId == "" {
		return nil, proto.ErrorEnums.ErrDepotNotExist
	}
	var depot Depot
	err := ds.depotMongo.Collection(proto.DatabaseName.DepotDataBaseName).FindOne(ctx, bson.M{"_id": depotId}).Decode(&depot)
	if err != nil {
		logx.Errorf("DepotServer|QueryDepotInfo|FindOne|err: %v", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, proto.ErrorEnums.ErrDepotNotExist
		}
		return nil, err
	}
	logx.Infof("DepotServer|QueryDepotInfo|depot: %s", conv.ToJsonWithoutError(depot))
	return &depot, nil
}
