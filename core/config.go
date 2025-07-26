package core

import (
	"github.com/BurntSushi/toml"
)

// Config 结构体定义了配置文件的结构
type Config struct {
	Group *string `toml:"group"`
	Port  *string `toml:"port"`
	Redis *string `toml:"redis"`
	Mongo *string `toml:"mongo"`
	S3    *S3     `toml:"s3"`
}

// S3 结构体定义了S3存储的配置
type S3 struct {
	Bucket    string `toml:"bucket"`
	Endpoint  string `toml:"endpoint"`
	AccessKey string `toml:"access_key"` // AccessKey for S3
	SecretKey string `toml:"secret_key"` // SecretKey for S3
	Region    string `toml:"region"`
}

// LoadConfig 从指定路径加载TOML配置文件
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
