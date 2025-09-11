package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dzjyyds666/mediaStorage/pkg"
	"net/url"

	"github.com/aws/smithy-go/ptr"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/mediaStorage/internal/config"
	"github.com/redis/go-redis/v9"
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

type BoxLogic struct {
	ctx    context.Context
	group  string
	boxRDB *redis.Client
}

func NewBoxLogic(ctx context.Context, conf *config.Config, dsServer *ds.DatabaseServer) *BoxLogic {
	boxRedis, ok := dsServer.GetRedis("box")
	if !ok {
		panic("redis [box] not found")
	}
	bs := &BoxLogic{
		ctx:    ctx,
		group:  ptr.ToString(conf.Group),
		boxRDB: boxRedis,
	}

	err := bs.StartCheck()
	if nil != err {
		panic(err)
	}
	return bs
}

// 构建box信息key
func (bs *BoxLogic) buildBoxInfoKey(id string) string {
	return fmt.Sprintf("media_storage:%s:box:%s:info", bs.group, id)
}

func (bs *BoxLogic) StartCheck() error {
	// 创建默认的box
	defaultBox := &Box{
		BoxId:   "default",
		BoxName: ptr.String("default"),
		DepotId: ptr.String("default"),
	}
	_, err := bs.CreateBox(bs.ctx, defaultBox)
	return err
}

// 创建盒子
func (bs *BoxLogic) CreateBox(ctx context.Context, info *Box) (*Box, error) {
	if len(info.BoxId) == 0 {
		info.BoxId = "bi_" + generateRandomString(8)
	}
	if info.DepotId == nil {
		info.DepotId = ptr.String("default")
	}
	raw, err := json.Marshal(info)
	if nil != err {
		logx.Errorf("BoxServer|CreateBox|json.Marshal|err: %v", err)
		return nil, err
	}

	boxInfoKey := bs.buildBoxInfoKey(info.BoxId)
	succ, err := bs.boxRDB.SetNX(ctx, boxInfoKey, raw, 0).Result()
	if nil != err {
		logx.Errorf("BoxServer|CreateBox|SetNx|err: %v", err)
		return nil, err
	}
	if !succ {
		info, err = bs.QueryBoxInfo(ctx, info.BoxId)
		if nil != err {
			logx.Errorf("BoxServer|CreateBox|QueryBoxInfo|boxId: %s|err: %v", info.BoxId, err)
			return nil, err
		}
		return info, nil
	}
	return info, nil
}

// 查询盒子的信息
func (bs *BoxLogic) QueryBoxInfo(ctx context.Context, boxId string) (*Box, error) {
	if len(boxId) == 0 {
		return nil, pkg.ErrorEnums.ErrBoxNotExist
	}
	key := bs.buildBoxInfoKey(boxId)
	result, err := bs.boxRDB.Get(ctx, key).Result()
	if err != nil {
		logx.Errorf("BoxServer|QueryBoxInfo|FindOne|boxId: %s|err: %v", boxId, err)
		if errors.Is(err, redis.Nil) {
			return nil, pkg.ErrorEnums.ErrBoxNotExist
		}
		return nil, err
	}
	var box Box
	err = json.Unmarshal([]byte(result), &box)
	if nil != err {
		logx.Errorf("BoxServer|QueryBoxInfo|json.Unmarshal|err: %v", err)
		return nil, err
	}
	return &box, nil
}
