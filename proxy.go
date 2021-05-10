package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// Proxy 代理结构
type Proxy struct {
	Ctx         context.Context
	proxyConfig *ProxyConfig
}

var _ http.Handler = &Proxy{}

// HeaderProxy http 代理
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	fmt.Println(req.URL.Host)

	// 禁止访问某个域名
	if p.proxyConfig == nil {
		p.proxyConfig = DefaultProxyConfig
	}
	if p.proxyConfig.BlcakURLs != nil {
		if _, ok := defaultBlackURLs[req.URL.Host]; ok {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}

	switch {
	case req.Method == http.MethodConnect:
		p.forwardTunnel(p.Ctx, req, rw)
	default:
		p.forwardHTTP(p.Ctx, req, rw)
	}

}

// forwardTunnel 隧道转发
func (p *Proxy) forwardTunnel(ctx context.Context, req *http.Request, rw http.ResponseWriter) {
	clientConn, err := hijacker(rw)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer clientConn.Close()

	targetAddr := req.URL.Host

	targetConn, err := net.DialTimeout("tcp", targetAddr, 5*time.Second) // 连接目标服务器超时时间
	if err != nil {
		log.Println(fmt.Errorf("%s - 隧道转发连接目标服务器失败: %s", req.URL.Host, err))
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	clientConn.SetDeadline(time.Now().Add(30 * time.Second)) // 客户端读写超时时间
	targetConn.SetDeadline(time.Now().Add(30 * time.Second)) // 目标服务器读写超时时间

	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")) // 隧道连接成功响应行
	if err != nil {
		log.Println(fmt.Errorf("%s - 隧道连接成功,通知客户端错误: %s", req.URL.Host, err))
		return
	}

	p.transfer(clientConn, targetConn)
}

// transfer 双向转发
func (p *Proxy) transfer(src net.Conn, dst net.Conn) {
	go func() {
		io.Copy(src, dst)
		src.Close()
		dst.Close()
	}()

	io.Copy(dst, src)
	dst.Close()
	src.Close()
}

// DoRequest 执行HTTP请求，并调用responseFunc处理response
func (p *Proxy) DoRequest(ctx context.Context, req *http.Request, responseFunc func(*http.Response, error)) {

	newReq := new(http.Request)
	*newReq = *req
	newReq.Header = CloneHeader(newReq.Header)
	removeConnectionHeaders(newReq.Header)
	for _, item := range hopHeaders {
		if newReq.Header.Get(item) != "" {
			newReq.Header.Del(item)
		}
	}
	resp, err := http.DefaultTransport.RoundTrip(newReq)
	if err == nil {
		removeConnectionHeaders(resp.Header)
		for _, h := range hopHeaders {
			resp.Header.Del(h)
		}
	}
	responseFunc(resp, err)
}

// forwardHTTP HTTP转发
func (p *Proxy) forwardHTTP(ctx context.Context, req *http.Request, rw http.ResponseWriter) {
	req.URL.Scheme = "http"

	p.DoRequest(ctx, req, func(resp *http.Response, err error) {
		if err != nil {
			log.Panicln(fmt.Errorf("%s - HTTP请求错误: , 错误: %s", req.URL, err))
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		CopyHeader(rw.Header(), resp.Header)
		rw.WriteHeader(resp.StatusCode)
		io.Copy(rw, resp.Body)
	})
}

// hijacker  获取底层连接
func hijacker(rw http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("web server不支持Hijacker")
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijacker错误: %s", err)
	}

	return conn, nil
}

// CopyHeader 浅拷贝Header
func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// CloneHeader 深拷贝Header
func CloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// CloneBody 拷贝Body
func CloneBody(b io.ReadCloser) (r io.ReadCloser, body []byte, err error) {
	if b == nil {
		return http.NoBody, nil, nil
	}
	body, err = ioutil.ReadAll(b)
	if err != nil {
		return http.NoBody, nil, err
	}
	r = ioutil.NopCloser(bytes.NewReader(body))

	return r, body, nil
}

var hopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func removeConnectionHeaders(h http.Header) {
	if c := h.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				h.Del(f)
			}
		}
	}
}
