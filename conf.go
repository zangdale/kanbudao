package main

import "os"

// 默认配置
var DefaultProxyConfig *ProxyConfig

func init() {
	// TODO: 读取配置信息
	isDefaut := os.Getenv("KANBUDAO_CONFIG_DEFAULT")
	if isDefaut != "" {
		// 读取默认配置文件
		DefaultProxyConfig = &ProxyConfig{
			BlcakURLs: defaultBlackURLs,
		}
	}
	// 读取 JSON 配置文件
	DefaultProxyConfig = &ProxyConfig{
		BlcakURLs: defaultBlackURLs,
	}
}

// ProxyConfig 代理的配置
type ProxyConfig struct {
	BlcakURLs map[string]struct{}
}
