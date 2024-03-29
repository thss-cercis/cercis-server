package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config 全局设置结构体
type Config struct {
	Server struct {
		Host   string
		Port   int
		Logger struct {
			Level uint32
		}
	}
	Redis struct {
		Host     string
		Port     int
		Username string
		Password string
		Database int
		Reset    bool
	}
	Postgres struct {
		Host     string
		Port     int
		User     string
		Password string
		Dbname   string
		Sslmode  string
		Timezone string
	}
	SMS struct {
		Region       string
		AccessKey    string
		Secret       string
		SignName     string
		TemplateCode string
	}
	Qiniu struct {
		AccessKey string
		SecretKey string
		Bucket    string
	}
}

var globalConfig *Config

// Init 初始化全局设置
func Init(filepath string) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &globalConfig)
	if err != nil {
		panic(err)
	}
}

// GetConfig 获得当前设置
func GetConfig() *Config {
	return globalConfig
}
