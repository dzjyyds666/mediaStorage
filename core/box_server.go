package core

import (
	"context"
	"errors"
	"net/url"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/mediaStorage/proto"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 箱子的结构
type Box struct {
	BoxId      string     `json:"box_id" bson:"_id"`
	BoxName    *string    `json:"box_name,omitempty" bson:"box_name,omitempty"`
	FileNumber *int64     `json:"file_number,omitempty" bson:"file_number,omitempty"`
	SpaceUsed  *int64     `json:"space_used,omitempty" bson:"space_used,omitempty"`
	MetaData   url.Values `json:"meta_data,omitempty" bson:"meta_data,omitempty"`
	Depot      *Depot     `json:"depot,omitempty" bson:"depot,omitempty"` // 属于哪个仓库
}

// 获取到仓库信息
func (b *Box) GetDepotId() string {
	if b.Depot == nil {
		return ""
	}
	return b.Depot.DepotId
}

type BoxServer struct {
	ctx      context.Context
	boxRDB   *redis.Client
	boxMongo *mongo.Database
}

func NewBoxServer(ctx context.Context, conf *Config, dsServer *ds.DatabaseServer) *BoxServer {
	boxRedis, ok := dsServer.GetRedis("box")
	if !ok {
		panic("redis [box] not found")
	}
	boxMongo, ok := dsServer.GetMongo("media_storage")
	if !ok {
		panic("mongo [media_storage] not found")
	}
	return &BoxServer{
		ctx:      ctx,
		boxRDB:   boxRedis,
		boxMongo: boxMongo,
	}
}

// 创建盒子
func (bs *BoxServer) CreateBox(ctx context.Context, box *Box) error {
	_, err := bs.boxMongo.Collection(proto.DatabaseName.BoxDataBaseName).InsertOne(ctx, box)
	if nil != err {
		logx.Errorf("BoxServer|CreateBox|InsertOne|err: %v", err)
		return err
	}
	logx.Infof("BoxServer|CreateBox|box: %s", conv.ToJsonWithoutError(box))
	return err
}

// 查询盒子的信息
func (bs *BoxServer) QueryBoxInfo(ctx context.Context, boxId string) (*Box, error) {
	var box Box
	err := bs.boxMongo.Collection(proto.DatabaseName.BoxDataBaseName).FindOne(ctx, bson.M{"_id": boxId}).Decode(&box)
	if err != nil {
		logx.Errorf("BoxServer|QueryBoxInfo|FindOne|err: %v", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, proto.ErrorEnums.ErrBoxNotExist
		}
		return nil, err
	}
	logx.Infof("BoxServer|QueryBoxInfo|box: %s", conv.ToJsonWithoutError(box))
	return &box, nil
}
