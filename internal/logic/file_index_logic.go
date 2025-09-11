package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/internal/config"
	"github.com/dzjyyds666/mediaStorage/pkg"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type FileOption func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error

// randFid 随机生成文件id
func randFid() string {
	return uuid.NewString()
}

type MediaFileInfo struct {
	Fid           string     `json:"fid" bson:"_id"` // 文件的fid
	FileName      string     `json:"file_name,omitempty" bson:"file_name,omitempty"`
	ContentMd5    *string    `json:"content_md5,omitempty" bson:"content_md5,omitempty"`
	ContentType   *string    `json:"content_type,omitempty" bson:"content_type,omitempty"`
	ContentLength *int64     `json:"content_length,omitempty" bson:"content_length,omitempty"`
	CreatedTs     *int64     `json:"created_ts,omitempty" bson:"created_ts,omitempty"`
	MetaData      url.Values `json:"meta_data,omitempty" bson:"meta_data,omitempty"`
	Uploader      *string    `json:"uploader,omitempty" bson:"uploader,omitempty"`
	Box           *Box       `json:"box,omitempty" bson:"box,omitempty"`

	r io.Reader // 文件的流
}

// BuildObjectKey 构建对象键
func (mfi *MediaFileInfo) BuildObjectKey() string {
	return path.Join(mfi.GetDepotId(), mfi.Box.BoxId, mfi.Fid)
}

func (mfi *MediaFileInfo) GetDepotId() string {
	return ptr.ToString(mfi.Box.DepotId)
}

type InitUpload struct {
	Fid           string     `json:"fid"`
	FileName      *string    `json:"file_name,omitempty"`
	ContentLength *int64     `json:"content_length,omitempty"`
	ContentMd5    *string    `json:"content_md5,omitempty"`
	ContentType   *string    `json:"content_type,omitempty"`
	Header        url.Values `json:"header,omitempty"`
	Uploader      *string    `json:"uploader,omitempty"`
	BoxId         *string    `json:"box_id,omitempty"`
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

type FileIndexLogic struct {
	ctx       context.Context
	group     string
	fileRedis *redis.Client
	s3Server  *S3Logic // s3 服务
}

// NewFileIndexLogic 创建文件索引服务
func NewFileIndexLogic(ctx context.Context, cfg *config.Config, dsServer *ds.DatabaseServer, s3Server *S3Logic, boxServ *BoxLogic, depotServ *DepotLogic) *FileIndexLogic {
	fileRedis, ok := dsServer.GetRedis("file")
	if !ok {
		panic("redis [file] not found")
	}
	return &FileIndexLogic{
		ctx:       ctx,
		group:     ptr.ToString(cfg.Group),
		fileRedis: fileRedis,
		s3Server:  s3Server,
	}
}

// 构建文件预备key
func (fl *FileIndexLogic) buildPrepareFileInfoKey(depotId, id string) string {
	return fmt.Sprintf("media_storage:%s:file:%s:%s:info:prepare", fl.group, depotId, id)
}

// 构建文件信息key
func (fl *FileIndexLogic) buildFileInfoKey(depotId, id string) string {
	return fmt.Sprintf("media_storage:%s:file:%s:%s:info", fl.group, depotId, id)
}

// 申请上传
func (fs *FileIndexLogic) CreatePrepareFileInfo(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
	for _, opt := range opts {
		opt(info)
	}
	raw, err := json.Marshal(info)
	if err != nil {
		logx.Errorf("FileIndexServer|CreatePrepareFileInfo|Marshal|err: %v", err)
		return err
	}

	// 把文件信息存储到redis中,1个小时之内进行上传
	_, err = fs.fileRedis.SetNX(ctx, fs.buildPrepareFileInfoKey(info.GetDepotId(), info.Fid), raw, time.Hour).Result()
	if err != nil {
		logx.Errorf("FileIndexServer|CreatePrepareFileInfo|Set|err: %v", err)
		return err
	}
	return nil
}

// 查询文件的prepare信息
func (fs *FileIndexLogic) QueryPrepareFileInfo(ctx context.Context, depotId, fid string) (*MediaFileInfo, error) {
	raw, err := fs.fileRedis.Get(ctx, fs.buildPrepareFileInfoKey(depotId, fid)).Bytes()
	if err != nil {
		logx.Errorf("FileIndexServer|QueryPerpareFileInfo|Get|err: %v", err)
		if errors.Is(err, redis.Nil) {
			return nil, pkg.ErrorEnums.ErrNoPrepareFileInfo
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

// 查询文件的信息
func (fs *FileIndexLogic) QueryFileInfo(ctx context.Context, depotId, fileId string) (*MediaFileInfo, error) {
	infoKey := fs.buildPrepareFileInfoKey(depotId, fileId)
	result, err := fs.fileRedis.Get(ctx, infoKey).Result()
	if nil != err {
		logx.Errorf("FileIndexServer|QueryFileInfo|Get|fileId: %s|err: %v", fileId, err)
		return nil, err
	}
	var info MediaFileInfo
	err = json.Unmarshal([]byte(result), &info)
	if err != nil {
		logx.Errorf("FileIndexServer|QueryFileInfo|FindOne|fileId: %s|err: %v", fileId, err)
		return nil, err
	}
	logx.Infof("FileIndexServer|QueryFileInfo|info|%s", conv.ToJsonWithoutError(info))
	return &info, nil
}

// 保存文件到s3
func (fs *FileIndexLogic) SaveFileData(ctx context.Context, info *MediaFileInfo, file io.Reader) error {
	info.r = file
	return fs.s3Server.SaveFileData(ctx, info)
}

// 完成文件上传
func (fs *FileIndexLogic) CompleteFileInfo(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
	// 把文件补全文件信息
	prepareInfo, err := fs.QueryPrepareFileInfo(ctx, info.GetDepotId(), info.Fid)
	if err != nil {
		logx.Errorf("FileIndexServer|CompleteUpload|QueryPerpareFileInfo|err: %v", err)
		return err
	}

	prepareInfo.CreatedTs = ptr.Int64(time.Now().Unix())
	prepareInfo.Box = info.Box
	prepareInfo.MetaData = info.MetaData

	infoKey := fs.buildFileInfoKey(info.GetDepotId(), info.Fid)
	_, err = fs.fileRedis.SetNX(ctx, infoKey, conv.ToJsonWithoutError(prepareInfo), time.Hour).Result()
	if err != nil {
		logx.Errorf("FileIndexServer|CompleteUpload|InsertOne|err: %v", err)
		return err
	}
	// 删除存储在redis中的数据
	err = fs.fileRedis.Del(ctx, fs.buildPrepareFileInfoKey(info.GetDepotId(), info.Fid)).Err()
	if err != nil {
		logx.Errorf("FileIndexServer|CompleteUpload|Del|err: %v", err)
	}
	return nil
}

func (fs *FileIndexLogic) SignFileUrl(ctx context.Context, info *MediaFileInfo) (string, error) {
	objectKey := info.BuildObjectKey()
	presignedURL, err := fs.s3Server.GetPresignedURL(ctx, objectKey)
	if err != nil {
		logx.Errorf("StorageCoreServer|SignGetFileUrl|GetPresignedURL|fid: %s|err: %s", info.Fid, err.Error())
		return "", err
	}
	return presignedURL, nil
}

// 申请文件上传
func (fs *FileIndexLogic) ApplyUpload(ctx context.Context, init *InitUpload, box *Box) (string, error) {
	// 生成文件的fid
	info := init.ToMediaFileInfo()
	info.Fid = randFid()
	init.Fid = info.Fid
	info.Box = box

	return info.Fid, do(
		fs.CreatePrepareFileInfo,
		func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			logx.Infof("StorageCoreServer|ApplyUpload|CreatePrepareFileInfo|info: %v", conv.ToJsonWithoutError(info))
			return nil
		},
	)(ctx, info)
}

// 文件直接上传
func (fs *FileIndexLogic) SingleUpload(ctx context.Context, box *Box, fid string, r io.Reader) error {
	// 查询文件的初始化上传信息
	prepareFileInfo, err := fs.QueryPrepareFileInfo(ctx, ptr.ToString(box.DepotId), fid)
	if err != nil {
		logx.Errorf("StorageCoreServer|SingleUpload|QueryPerpareFileInfo|boxId: %s|fid: %s|err: %s", box.BoxId, fid, err.Error())
		return err
	}
	return do(
		// 开始上传文件
		func(ctx context.Context, info *MediaFileInfo, opts ...func(*MediaFileInfo) *MediaFileInfo) error {
			err := fs.SaveFileData(ctx, info, r)
			if nil != err {
				logx.Errorf("StorageCoreServer|SingleUpload|SaveFileData|boxId: %s|fid: %s|err: %s", box.BoxId, fid, err.Error())
				return err
			}
			return nil
		},
		// 完成上传之后的文件信息构建
		fs.CompleteFileInfo,
	)(ctx, prepareFileInfo)
}
