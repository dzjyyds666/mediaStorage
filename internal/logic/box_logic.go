package logic

import (
	"context"
	"errors"
	"net/url"

	"github.com/aws/smithy-go/ptr"
	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/mediaStorage/internal/config"
	"github.com/dzjyyds666/mediaStorage/pkg"
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
	DepotId    *string    `json:"depot_id,omitempty" bson:"depot_id,omitempty"`
}

type BoxServer struct {
	ctx      context.Context
	boxRDB   *redis.Client
	boxMongo *mongo.Database
}

func NewBoxServer(ctx context.Context, conf *config.Config, dsServer *ds.DatabaseServer) *BoxServer {
	boxRedis, ok := dsServer.GetRedis("box")
	if !ok {
		panic("redis [box] not found")
	}
	boxMongo, ok := dsServer.GetMongo("media_storage")
	if !ok {
		panic("mongo [media_storage] not found")
	}
	bs := &BoxServer{
		ctx:      ctx,
		boxRDB:   boxRedis,
		boxMongo: boxMongo,
	}

	err := bs.StartCheck()
	if nil != err {
		panic(err)
	}
	return bs
}

func (bs *BoxServer) StartCheck() error {
	// 创建默认的box
	defaultBox := &Box{
		BoxId:   "default",
		BoxName: ptr.String("default"),
		DepotId: ptr.String("default"),
	}
	return bs.CreateBox(bs.ctx, defaultBox)
}

// 创建盒子
func (bs *BoxServer) CreateBox(ctx context.Context, info *Box) error {
	if len(info.BoxId) == 0 {
		info.BoxId = "bi_" + generateRandomString(8)
	}
	if info.DepotId == nil {
		info.DepotId = ptr.String("default")
	}
	_, err := bs.boxMongo.Collection(pkg.DatabaseName.BoxDataBaseName).InsertOne(ctx, info)
	if nil != err {
		logx.Errorf("BoxServer|CreateBox|InsertOne|err: %v", err)
		if mongo.IsDuplicateKeyError(err) {
			// 存在即不插入
			return nil
		}
		return err
	}
	logx.Infof("BoxServer|CreateBox|box: %s", conv.ToJsonWithoutError(info))
	return err
}

// 查询盒子的信息
func (bs *BoxServer) QueryBoxInfo(ctx context.Context, boxId string) (*Box, error) {
	var box Box
	err := bs.boxMongo.Collection(pkg.DatabaseName.BoxDataBaseName).FindOne(ctx, bson.M{"_id": boxId}).Decode(&box)
	if err != nil {
		logx.Errorf("BoxServer|QueryBoxInfo|FindOne|boxId: %s|err: %v", boxId, err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrorEnums.ErrBoxNotExist
		}
		return nil, err
	}
	logx.Infof("BoxServer|QueryBoxInfo|box|%s", conv.ToJsonWithoutError(box))
	return &box, nil
}
