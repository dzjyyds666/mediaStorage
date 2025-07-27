package core

import (
	"context"
	"io"
	"net/url"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type MediaFileInfo struct {
	FileName      string     `json:"file_name"`
	ContentMd5    *string    `json:"content_md5"`
	ContentType   *string    `json:"content_type"`
	ContentLength *int64     `json:"content_length"`
	CreatedTs     *int64     `json:"created_ts"`
	Header        url.Values `json:"header"`

	r io.Reader // 文件的流

}

type FileIndexServer struct {
	ctx       context.Context
	fileRedis *redis.Client
	fileMongo *mongo.Client
	s3Server  *S3Server // s3 服务
}

// NewFileIndexServer 创建文件索引服务
func NewFileIndexServer(ctx context.Context, cfg *Config, dsServer *ds.DatabaseServer, s3Server *S3Server) *FileIndexServer {
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
