package core

import (
	"context"
	"encoding/json"
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
	filePrepareKey = "file:upload:%s:prepare"
)

func buildFilePrepareKey(fid string) string {
	return fmt.Sprintf(filePrepareKey, fid)
}

type MediaFileInfo struct {
	Fid           string     `json:"fid"`
	FileName      string     `json:"file_name,omitempty"`
	ContentMd5    *string    `json:"content_md5,omitempty"`
	ContentType   *string    `json:"content_type,omitempty"`
	ContentLength *int64     `json:"content_length,omitempty"`
	CreatedTs     *int64     `json:"created_ts,omitempty"`
	Header        url.Values `json:"header,omitempty"`
	Box           *string    `json:"box,omitempty"`
	Slice         *string    `json:"slice,omitempty"`

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
func (fs *FileIndexServer) ApplyUpload(ctx context.Context, uploadInfo *InitUpload) (string, error) {
	if uploadInfo.FileName == nil {
		return "", proto.ErrorEnums.ErrFileNameCanNotBeEmpty
	}
	if uploadInfo.FileSize == nil {
		return "", proto.ErrorEnums.ErrFileSizeCanNotBeZero
	}
	if uploadInfo.ContentType == nil {
		return "", proto.ErrorEnums.ErrFileTypeCanNotBeEmpty
	}
	// 把文件信息存储到redis中
	fileInfo := &MediaFileInfo{
		FileName:      *uploadInfo.FileName,
		ContentMd5:    uploadInfo.ContentMd5,
		ContentType:   uploadInfo.ContentType,
		ContentLength: uploadInfo.FileSize,
	}
	raw, err := json.Marshal(fileInfo)
	if err != nil {
		logx.Errorf("FileIndexServer|ApplyUpload|Marshal|err: %v", err)
		return "", err
	}

	fid := fs.randFid()

	// 把文件信息存储到redis中,1个小时之内进行上传
	err = fs.fileRedis.Set(ctx, buildFilePrepareKey(fid), raw, time.Hour).Err()
	if err != nil {
		logx.Errorf("FileIndexServer|ApplyUpload|Set|err: %v", err)
		return "", err
	}
	return fid, nil
}

// randFid 随机生成文件id
func (fs *FileIndexServer) randFid() string {
	return "v1-" + uuid.NewString()
}
