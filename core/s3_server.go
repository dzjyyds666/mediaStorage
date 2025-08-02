package core

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Server struct {
	ctx    context.Context
	client *s3.Client // s3客户端
}

// 创建s3服务，直接操作s3
func NewS3Server(ctx context.Context, cfg *Config) *S3Server {
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
	return &S3Server{
		ctx:    ctx,
		client: s3Client,
	}
}
