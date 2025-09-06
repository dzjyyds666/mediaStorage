package logic

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/Allspark-go/logx"
	myconfig "github.com/dzjyyds666/mediaStorage/internal/config"
)

type S3Server struct {
	ctx    context.Context
	bucket string
	client *s3.Client // s3客户端
}

// 创建s3服务，直接操作s3
func NewS3Server(ctx context.Context, cfg *myconfig.Config) *S3Server {
	// 创建s3客户端
	s3Cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.S3.Region),
		config.WithBaseEndpoint(cfg.S3.Endpoint),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, "")),
	)
	if err != nil {
		panic(err)
	}

	s3Client := s3.NewFromConfig(s3Cfg)

	// 检查bucket是否存在
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.S3.Bucket),
	})
	if err != nil {
		// bucket不存在，创建新的bucket
		_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(cfg.S3.Bucket),
		})
		if err != nil {
			panic(err)
		}
	}

	return &S3Server{
		ctx:    ctx,
		bucket: cfg.S3.Bucket,
		client: s3Client,
	}
}

// SaveFileData 保存文件信息到s3
func (ss *S3Server) SaveFileData(ctx context.Context, info *MediaFileInfo) error {
	if info.r == nil {
		return errors.New("file data is nil")
	}
	objKey := info.BuildObjectKey()

	_, err := ss.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(ss.bucket),
		Key:         aws.String(objKey),
		Body:        info.r,
		ContentType: info.ContentType,
	})
	if nil != err {
		logx.Errorf("S3Server|SaveFileData|PutObject|err: %v", err)
		return err
	}
	return nil
}

// 获取s3的访问预签名url
func (ss *S3Server) GetPresignedURL(ctx context.Context, objectKey string) (string, error) {
	presignClient := s3.NewPresignClient(ss.client)
	presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(ss.bucket),
		Key:    aws.String(objectKey),
	})
	if nil != err {
		logx.Errorf("S3Server|GetPresignedURL|GetPresignedURL|err: %v", err)
		return "", err
	}
	return presignedURL.URL, nil
}
