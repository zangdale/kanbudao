package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getbuguai/gox/xpath"
	"github.com/getbuguai/gox/xtools"
)

// 默认配置
var DefaultProxyConfig *ProxyConfig

const (
	defaultConfigFileName   string = "config.json"
	defaultConfigServerPort uint64 = 9977
)

func init() {
	//  读取配置信息
	fileDir, _, err := xpath.ExecPath()
	if err != nil {
		log.Fatal(err)
	}

	configFile := filepath.Join(fileDir, defaultConfigFileName)

	var conf = new(config)
	bUrls := defaultBlackURLs
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		conf, err = LoadConfig(configFile)
		if err != nil {
			log.Println(err)
		} else {
			bUrls = conf.blcakURLs()
		}
	}

	DefaultProxyConfig = &ProxyConfig{
		ServerPort: xtools.IF(conf == nil || conf.Port == 0 || conf.Port > 65535, defaultConfigServerPort, conf.Port).(uint64),
		BlcakURLs:  bUrls,
	}
}

// ProxyConfig 代理的配置
type ProxyConfig struct {
	ServerPort uint64
	BlcakURLs  map[string]struct{}
}

// config 配置文件中的内容
type config struct {
	Port uint64 `json:"port,omitempty"`
	Urls []struct {
		Url  string `json:"url"`
		Port uint64 `json:"port,omitempty"`
	} `json:"urls"`
}

// LoadConfig 加载配置文件
func LoadConfig(filePath string) (c *config, err error) {
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	conf := new(config)
	err = json.Unmarshal(body, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// blcakURLs 获取域名黑名单
func (conf *config) blcakURLs() (blcakURLs map[string]struct{}) {

	blcakURLs = make(map[string]struct{})

	for _, u := range conf.Urls {
		u.Url = strings.ToLower(u.Url)
		u.Url = strings.TrimPrefix(strings.TrimPrefix(u.Url, "https://"), "http://")
		if u.Port == 0 || u.Port > 65535 {
			u.Port = 443
		}

		blcakURLs[fmt.Sprintf("%s:%d", u.Url, u.Port)] = empty
	}

	return blcakURLs
}
