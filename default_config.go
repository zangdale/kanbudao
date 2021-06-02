package main

var empty = struct{}{}

var DefaultMsg = []byte("Hello Black URL !")

var defaultBlackURLs = map[string]struct{}{
	"script.hotjar.com:443":   empty,
	"static.hotjar.com:443":   empty,
	"insights.hotjar.com:443": empty,
	"ws3.hotjar.com:443":      empty,
	"vars.hotjar.com:443":     empty,
	"sp0.baidu.com:443":       empty,
	"hm.baidu.com:443":        empty}
