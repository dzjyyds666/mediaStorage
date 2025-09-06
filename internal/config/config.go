package config

import (
	"github.com/BurntSushi/toml"
	"github.com/dzjyyds666/Allspark-go/ds"
)

// Config 结构体定义了配置文件的结构
type Config struct {
	Group  *string `toml:"group"`
	Port   *string `toml:"port"`
	S3     *S3     `toml:"s3"`
	Server *Server `toml:"server"`
	Admin  *Admin  `toml:"admin"`
}

type Admin struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// S3 结构体定义了S3存储的配置
type S3 struct {
	Bucket    string `toml:"bucket"`
	Endpoint  string `toml:"endpoint"`
	AccessKey string `toml:"access_key"` // AccessKey for S3
	SecretKey string `toml:"secret_key"` // SecretKey for S3
	Region    string `toml:"region"`
}

type Server struct {
	DBConfig   *ds.DsConfig `toml:"ds_config"`   // 数据库配置
	Jwt        *Jwt         `toml:"jwt"`         // 服务端jwt
	ConsoleJwt *Jwt         `toml:"console_jwt"` // 控制台jwt
}

type Jwt struct {
	Secret string `toml:"secret"`
	Expire int64  `toml:"expire"`
}

// LoadConfig 从指定路径加载TOML配置文件
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
