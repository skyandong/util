package conf

import (
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"path"
	"strings"
)

type ConfigLogLevel int

const (
	LevelTrace ConfigLogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
	LevelFatal
)

var SupportedExtensions = []string{"json", "toml", "yaml", "yml", "properties", "props", "prop", "hcl"}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func SetLogLevel(level ConfigLogLevel) {
	// 打印 viper 配置文件加载过程
	jwwLogLevel := jww.Threshold(level)
	jww.SetStdoutThreshold(jwwLogLevel)
}

func init() {
	SetLogLevel(LevelInfo)
}

func New() *XConfig {
	xViper := viper.New()
	return &XConfig{xViper}
}

type XConfig struct {
	viper *viper.Viper
}

func (conf *XConfig) SetConfigName(in string) {
	conf.viper.SetConfigName(in)
}

func (conf *XConfig) SetConfigType(in string) {
	conf.viper.SetConfigType(in)
}

func (conf *XConfig) AddConfigPath(in string) {
	conf.viper.AddConfigPath(in)
}

func (conf *XConfig) ReadInConfig() error {
	return conf.viper.ReadInConfig()
}

func (conf *XConfig) Unmarshal(rawVal interface{}) error {
	return conf.viper.Unmarshal(rawVal)
}

// confPath 配置文件路径  如：/datarver/pro/conf.yaml
// rawVal 配置文件映射的对象
func LoadConfig(confPath string, rawVal interface{}) (err error) {
	confPath = strings.Replace(confPath, "\\", "/", -1)
	fileDir := path.Dir(confPath)
	fileFullName := path.Base(confPath)
	fileExtension := path.Ext(fileFullName)
	if fileExtension == "" {
		fileExtension = ".yaml"
	}
	fileType := fileExtension[1:]
	if !stringInSlice(fileType, SupportedExtensions) {
		return viper.UnsupportedConfigError(fileType)
	}
	nameLen := len(fileFullName) - len(fileExtension)
	configName := fileFullName[:nameLen]
	conf := New()
	conf.SetConfigName(configName) // 配置文件的名字
	conf.SetConfigType(fileType)   // 配置文件的类型
	conf.AddConfigPath(fileDir)    // 配置文件的路径

	if err = conf.ReadInConfig(); err != nil {
		return err
	}

	if err = conf.Unmarshal(rawVal); err != nil {
		panic(fmt.Errorf("unable to decode into struct：  %s \n", err))
	}

	return nil
}
