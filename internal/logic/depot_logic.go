package logic

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"

	"github.com/aws/smithy-go/ptr"
	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/mediaStorage/internal/config"
	"github.com/dzjyyds666/mediaStorage/pkg"
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

// generateRandomString 生成指定长度的随机字符串
// length: 字符串长度
// charset: 字符集，如果为空则使用默认字符集（数字+字母）
func generateRandomString(length int, charset ...string) string {
	if length <= 0 {
		return ""
	}

	// 默认字符集：数字+大小写字母
	defaultCharset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	charSet := defaultCharset

	if len(charset) > 0 && charset[0] != "" {
		charSet = charset[0]
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		if err != nil {
			// 如果随机数生成失败，回退到UUID方式
			uuidStr := uuid.NewString() + uuid.NewString() // 确保有足够长度
			if len(uuidStr) >= length {
				return uuidStr[:length]
			}
			return uuidStr
		}
		result[i] = charSet[num.Int64()]
	}
	return string(result)
}

type Depot struct {
	DepotId        string     `json:"depot_id" bson:"_id"`
	DepotName      *string    `json:"depot_name,omitempty" bson:"depot_name,omitempty"`
	Permission     *string    `json:"permission,omitempty" bson:"permission,omitempty"`
	PermissionHook *string    `json:"permission_hook,omitempty" bson:"permission_hook,omitempty"` // 权限钩子,是一个url类型的，可以是webhook，也可以是redis
	MetaData       url.Values `json:"meta_data,omitempty" bson:"meta_data,omitempty"`             // 元数据
}

// 切片服务，文件存储分为两部分 桶 => 仓库 => 箱子 => file
type DepotLogic struct {
	ctx        context.Context
	depotRDB   *redis.Client
	depotMongo *mongo.Database
	boxServ    *BoxLogic
}

// 仓库
func NewDepotLogic(ctx context.Context, cfg *config.Config, dsServer *ds.DatabaseServer, boxServer *BoxLogic) *DepotLogic {
	depotRedis, ok := dsServer.GetRedis("depot")
	if !ok {
		panic("redis [depot] not found")
	}
	depotMongo, ok := dsServer.GetMongo("media_storage")
	if !ok {
		panic("mongo [media_storage] not found")
	}

	ds := &DepotLogic{
		ctx:        ctx,
		depotRDB:   depotRedis,
		depotMongo: depotMongo,
		boxServ:    boxServer,
	}

	err := ds.StartCheck()
	if nil != err {
		panic(err)
	}

	return ds
}

// 启动检查
func (ds *DepotLogic) StartCheck() error {
	// 创建默认的depot
	defaultDepot := &Depot{
		DepotId:    "default",
		DepotName:  ptr.String("default"),
		Permission: ptr.String(DepotPermissions.Public),
	}
	return ds.CreateDepot(ds.ctx, defaultDepot)
}

// 创建仓库
func (ds *DepotLogic) CreateDepot(ctx context.Context, info *Depot) error {
	if len(info.DepotId) == 0 {
		info.DepotId = "di_" + generateRandomString(8)
	}
	if info.Permission == nil {
		info.Permission = ptr.String(DepotPermissions.Public)
	}

	_, err := ds.depotMongo.Collection(pkg.DatabaseName.DepotDataBaseName).InsertOne(ctx, info)
	if nil != err {
		if mongo.IsDuplicateKeyError(err) {
			// 存在了就不插入新的
			return nil
		}
		logx.Errorf("DepotServer|CreateDepot|InsertOne|err: %v", err)
		return err
	}
	logx.Infof("DepotServer|CreateDepot|success|depot_info: %s", conv.ToJsonWithoutError(info))
	return nil
}

// 查询仓库信息
func (ds *DepotLogic) QueryDepotInfo(ctx context.Context, depotId string) (*Depot, error) {
	if depotId == "" {
		return nil, pkg.ErrorEnums.ErrDepotNotExist
	}
	var depot Depot
	err := ds.depotMongo.Collection(pkg.DatabaseName.DepotDataBaseName).FindOne(ctx, bson.M{"_id": depotId}).Decode(&depot)
	if err != nil {
		logx.Errorf("DepotServer|QueryDepotInfo|FindOne|err: %v", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrorEnums.ErrDepotNotExist
		}
		return nil, err
	}
	logx.Infof("DepotServer|QueryDepotInfo|depot: %s", conv.ToJsonWithoutError(depot))
	return &depot, nil
}

func do(funcs ...FileOption) FileOption {
	return func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
		for _, f := range funcs {
			if err := f(ctx, info, opts...); err != nil {
				return err
			}
		}
		return nil
	}
}
