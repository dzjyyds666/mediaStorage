package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/proto"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func buildFilePrepareKey(depot, box, fid string) string {
	return fmt.Sprintf("file:%s:%s:%s:prepare", depot, box, fid)
}

type FileOption func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error

type MediaFileInfo struct {
	Fid           string     `json:"fid" bson:"_id"` // 文件的fid
	FileName      string     `json:"file_name,omitempty" bson:"file_name,omitempty"`
	ContentMd5    *string    `json:"content_md5,omitempty" bson:"content_md5,omitempty"`
	ContentType   *string    `json:"content_type,omitempty" bson:"content_type,omitempty"`
	ContentLength *int64     `json:"content_length,omitempty" bson:"content_length,omitempty"`
	CreatedTs     *int64     `json:"created_ts,omitempty" bson:"created_ts,omitempty"`
	MetaData      url.Values `json:"meta_data,omitempty" bson:"meta_data,omitempty"`
	Box           *string    `json:"box,omitempty" bson:"box,omitempty"`
	Slice         *string    `json:"slice,omitempty" bson:"slice,omitempty"`
	Uploader      *string    `json:"uploader,omitempty" bson:"uploader,omitempty"`
	BoxInfo       *Box       `json:"box_info,omitempty" bson:"box_info,omitempty"` // 属于哪个box

	r io.Reader // 文件的流
}

func (mfi *MediaFileInfo) GetDepot() *Depot {
	return mfi.BoxInfo.Depot
}

type InitUpload struct {
	Fid           string     `json:"fid"`
	FileName      *string    `json:"file_name,omitempty"`
	ContentLength *int64     `json:"content_length,omitempty"`
	ContentMd5    *string    `json:"content_md5,omitempty"`
	ContentType   *string    `json:"content_type,omitempty"`
	Header        url.Values `json:"header,omitempty"`
}

// 转换为媒体文件信息
func (i *InitUpload) ToMediaFileInfo() *MediaFileInfo {
	return &MediaFileInfo{
		FileName:      ptr.ToString(i.FileName),
		ContentLength: i.ContentLength,
		ContentMd5:    i.ContentMd5,
		ContentType:   i.ContentType,
		MetaData:      i.Header,
	}
}

type FileIndexServer struct {
	ctx       context.Context
	fileRedis *redis.Client
	fileMongo *mongo.Database
	s3Server  *S3Server // s3 服务
	boxServ   *BoxServer
	depotServ *DepotServer
}

// NewFileIndexServer 创建文件索引服务
func NewFileIndexServer(ctx context.Context, cfg *Config, dsServer *ds.DatabaseServer, s3Server *S3Server, boxServ *BoxServer, depotServ *DepotServer) *FileIndexServer {
	fileRedis, ok := dsServer.GetRedis("file")
	if !ok {
		panic("redis [file] not found")
	}
	fileMongo, ok := dsServer.GetMongo("media_storage")
	if !ok {
		panic("mongo [media_storage] not found")
	}

	return &FileIndexServer{
		ctx:       ctx,
		fileRedis: fileRedis,
		fileMongo: fileMongo,
		s3Server:  s3Server,
	}
}

// 申请上传
func (fs *FileIndexServer) CreatePrepareFileInfo(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
	raw, err := json.Marshal(info)
	if err != nil {
		logx.Errorf("FileIndexServer|CreatePrepareFileInfo|Marshal|err: %v", err)
		return err
	}
	// 把文件信息存储到redis中,1个小时之内进行上传
	ok, err := fs.fileRedis.SetNX(ctx, buildFilePrepareKey(info.BoxInfo.GetDepotId(), info.BoxInfo.BoxId, info.Fid), raw, time.Hour).Result()
	if err != nil {
		logx.Errorf("FileIndexServer|CreatePrepareFileInfo|Set|err: %v", err)
		return err
	}
	if !ok {
		logx.Errorf("FileIndexServer|CreatePrepareFileInfo|SetNX|fid: %s|err: %v", info.Fid, proto.ErrorEnums.ErrFileExist)
		return proto.ErrorEnums.ErrFileExist
	}
	return nil
}

// randFid 随机生成文件id
func (fs *FileIndexServer) randFid() string {
	return "v1-" + uuid.NewString()
}

// 查询文件的prepare信息
func (fs *FileIndexServer) QueryPerpareFileInfo(ctx context.Context, depotId, boxId, fid string) (*MediaFileInfo, error) {
	raw, err := fs.fileRedis.Get(ctx, buildFilePrepareKey(depotId, boxId, fid)).Bytes()
	if err != nil {
		logx.Errorf("FileIndexServer|QueryPerpareFileInfo|Get|err: %v", err)
		if errors.Is(err, redis.Nil) {
			return nil, proto.ErrorEnums.ErrNoPrepareFileInfo
		}
		return nil, err
	}
	var info MediaFileInfo
	err = json.Unmarshal(raw, &info)
	if err != nil {
		logx.Errorf("FileIndexServer|QueryPerpareFileInfo|Unmarshal|err: %v", err)
	}
	return &info, nil
}

// 文件上传完毕，文件信息写入到mongo
func (fs *FileIndexServer) CreateFileInfo(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
	for _, opt := range opts {
		opt(info)
	}
	_, err := fs.fileMongo.Collection(proto.DatabaseName.FileDataBaseName).InsertOne(ctx, info)
	if err != nil {
		logx.Errorf("FileIndexServer|CreateFileInfo|InsertOne|err: %v", err)
		return err
	}
	logx.Infof("FileIndexServer|CreateFileInfo|info: %s", conv.ToJsonWithoutError(info))
	return nil
}

// 查询文件的信息
func (fs *FileIndexServer) GetFileInfo(ctx context.Context, fileId string) (*MediaFileInfo, error) {
	var info MediaFileInfo
	err := fs.fileMongo.Collection(proto.DatabaseName.FileDataBaseName).FindOne(ctx, bson.M{"fid": fileId}).Decode(&info)
	if err != nil {
		logx.Errorf("FileIndexServer|GetFileInfo|FindOne|err: %v", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, proto.ErrorEnums.ErrFileNotExist
		}
		return nil, err
	}
	logx.Infof("FileIndexServer|GetFileInfo|info: %s", conv.ToJsonWithoutError(info))
	return &info, nil
}
