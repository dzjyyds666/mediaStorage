package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/mediaStorage/proto"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	filePrepareKey = "file:%s:%s:%s:prepare"
)

func buildFilePrepareKey(depot, box, fid string) string {
	return fmt.Sprintf(filePrepareKey, depot, box, fid)
}

type FileOption func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error

type MediaFileInfo struct {
	Fid           string     `json:"fid"` // 文件的fid
	FileName      string     `json:"file_name,omitempty"`
	ContentMd5    *string    `json:"content_md5,omitempty"`
	ContentType   *string    `json:"content_type,omitempty"`
	ContentLength *int64     `json:"content_length,omitempty"`
	CreatedTs     *int64     `json:"created_ts,omitempty"`
	MetaData      url.Values `json:"meta_data,omitempty"`
	Box           *string    `json:"box,omitempty"`
	Slice         *string    `json:"slice,omitempty"`
	Uploader      *string    `json:"uploader,omitempty"`
	BoxInfo       *Box       `json:"box_info,omitempty"` // 属于哪个box

	r io.Reader // 文件的流
}

type InitUpload struct {
	FileName    *string    `json:"file_name,omitempty"`
	FileSize    *int64     `json:"file_size,omitempty"`
	ContentMd5  *string    `json:"content_md5,omitempty"`
	ContentType *string    `json:"content_type,omitempty"`
	Header      url.Values `json:"header,omitempty"`
}

type FileIndexServer struct {
	ctx       context.Context
	fileRedis *redis.Client
	fileMongo *mongo.Client
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
	err = fs.fileRedis.Set(ctx, buildFilePrepareKey(info.Fid), raw, time.Hour).Err()
	if err != nil {
		logx.Errorf("FileIndexServer|CreatePrepareFileInfo|Set|err: %v", err)
		return err
	}
	return nil
}

// randFid 随机生成文件id
func (fs *FileIndexServer) randFid() string {
	return "v1-" + uuid.NewString()
}

// 查询文件的prepare信息
func (fs *FileIndexServer) QueryPerpareFileInfo(ctx context.Context, fid string) (*MediaFileInfo, error) {
	raw, err := fs.fileRedis.Get(ctx, buildFilePrepareKey(fid)).Bytes()
	if err != nil {
		logx.Errorf("FileIndexServer|QueryPerpareFileInfo|Get|err: %v", err)
		if errors.Is(err, redis.Nil) {
			return nil, proto.ErrorEnums.ErrNoPrepareFileInfo
		}
		return nil, err
	}
	var info *MediaFileInfo
	err = json.Unmarshal(raw, &info)
	if err != nil {
		logx.Errorf("FileIndexServer|QueryPerpareFileInfo|Unmarshal|err: %v", err)
	}
	return info, nil
}
