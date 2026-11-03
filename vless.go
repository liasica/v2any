package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// proxy 是 mihomo 的 VLESS 节点配置，字段顺序即输出顺序
type proxy struct {
	Type              string       `yaml:"type"`
	Name              string       `yaml:"name"`
	Server            string       `yaml:"server"`
	Port              int          `yaml:"port"`
	UUID              string       `yaml:"uuid"`
	UDP               bool         `yaml:"udp"`
	TLS               bool         `yaml:"tls"`
	Flow              string       `yaml:"flow,omitempty"`
	ClientFingerprint string       `yaml:"client-fingerprint,omitempty"`
	SkipCertVerify    bool         `yaml:"skip-cert-verify"`
	XUDP              bool         `yaml:"xudp,omitempty"`
	PacketAddr        bool         `yaml:"packet-addr,omitempty"`
	RealityOpts       *realityOpts `yaml:"reality-opts,omitempty"`
	Network           string       `yaml:"network"`
	Encryption        string       `yaml:"encryption"`
	ServerName        string       `yaml:"servername,omitempty"`
	ALPN              []string     `yaml:"alpn,omitempty"`
	Fingerprint       string       `yaml:"fingerprint,omitempty"`
	HTTPOpts          *httpOpts    `yaml:"http-opts,omitempty"`
	H2Opts            *h2Opts      `yaml:"h2-opts,omitempty"`
	WSOpts            *wsOpts      `yaml:"ws-opts,omitempty"`
	GRPCOpts          *grpcOpts    `yaml:"grpc-opts,omitempty"`
	XHTTPOpts         *xhttpOpts   `yaml:"xhttp-opts,omitempty"`
}

type realityOpts struct {
	PublicKey             string `yaml:"public-key"`
	ShortID               string `yaml:"short-id,omitempty"`
	SpiderX               string `yaml:"_spider-x,omitempty"`
	SupportX25519MLKEM768 bool   `yaml:"support-x25519mlkem768,omitempty"`
}

type httpOpts struct {
	Method  string              `yaml:"method,omitempty"`
	Path    []string            `yaml:"path"`
	Headers map[string][]string `yaml:"headers,omitempty"`
}

type h2Opts struct {
	Host []string `yaml:"host,omitempty"`
	Path string   `yaml:"path"`
}

type wsOpts struct {
	Path                string            `yaml:"path"`
	Headers             map[string]string `yaml:"headers,omitempty"`
	MaxEarlyData        int               `yaml:"max-early-data,omitempty"`
	EarlyDataHeaderName string            `yaml:"early-data-header-name,omitempty"`
	HTTPUpgrade         bool              `yaml:"v2ray-http-upgrade,omitempty"`
	HTTPUpgradeFastOpen bool              `yaml:"v2ray-http-upgrade-fast-open,omitempty"`
}

type grpcOpts struct {
	ServiceName string `yaml:"grpc-service-name,omitempty"`
}

type xhttpOpts struct {
	Path string `yaml:"path"`
	Host string `yaml:"host,omitempty"`
	Mode string `yaml:"mode,omitempty"`
}

// parseLink 把 vless 分享链接转换为 mihomo 节点配置
// 链接格式: https://github.com/XTLS/Xray-core/discussions/716
func parseLink(link string) (*proxy, error) {
	location, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if location.Scheme != "vless" {
		return nil, fmt.Errorf("不是 vless 链接: %q", link)
	}

	// url.Values.Get 会吞掉解析错误，这里显式解析以便发现损坏的链接
	var query url.Values
	query, err = url.ParseQuery(location.RawQuery)
	if err != nil {
		return nil, err
	}

	// 校验必填的节点信息，UUID 只判空: mihomo 也接受非标准格式的字符串
	server := location.Hostname()
	if server == "" {
		return nil, errors.New("缺少服务器地址")
	}
	var port int
	port, err = strconv.Atoi(location.Port())
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("端口无效: %q", location.Port())
	}
	uuid := location.User.Username()
	if uuid == "" {
		return nil, errors.New("缺少 UUID")
	}

	// 备注为空或与服务器地址相同时没有区分作用，改用「地址:端口」
	name := location.Fragment
	if name == "" || name == server {
		name = location.Host
	}

	result := &proxy{
		Type:           "vless",
		Name:           name,
		Server:         server,
		Port:           port,
		UUID:           uuid,
		UDP:            true,
		Flow:           strings.ToLower(query.Get("flow")),
		SkipCertVerify: queryBool(query, "allowInsecure") || queryBool(query, "insecure"),
		Encryption:     queryOrDefault(query, "encryption", "none"),
		ServerName:     query.Get("sni"),
	}

	// UDP 的封装方式，链接未指定时默认使用 xudp
	switch query.Get("packetEncoding") {
	case "none":
	case "packet":
		result.PacketAddr = true
	default:
		result.XUDP = true
	}

	if err = applySecurity(result, query); err != nil {
		return nil, err
	}
	if err = applyTransport(result, query); err != nil {
		return nil, err
	}

	return result, nil
}

// applySecurity 处理传输层安全参数，无法识别的取值直接报错，避免静默退化为明文连接
func applySecurity(result *proxy, query url.Values) error {
	security := strings.ToLower(query.Get("security"))
	switch security {
	case "", "none":
		return nil
	case "tls", "xtls", "reality":
		result.TLS = true
		result.ClientFingerprint = queryOrDefault(query, "fp", "chrome")
		result.ALPN = splitList(query.Get("alpn"))
		result.Fingerprint = query.Get("pcs")
	default:
		return fmt.Errorf("不支持的 security: %q", security)
	}

	if security != "reality" {
		return nil
	}

	publicKey := query.Get("pbk")
	if publicKey == "" {
		return errors.New("reality 链接缺少 pbk")
	}
	result.RealityOpts = &realityOpts{
		PublicKey:             publicKey,
		ShortID:               query.Get("sid"),
		SpiderX:               query.Get("spx"),
		SupportX25519MLKEM768: queryBool(query, "support-x25519mlkem768"),
	}

	return nil
}

// applyTransport 把链接的传输方式映射为 mihomo 的 network 和对应选项
func applyTransport(result *proxy, query url.Values) error {
	network := strings.ToLower(query.Get("type"))
	switch network {
	case "", "tcp", "raw":
		result.Network = "tcp"

		// tcp 的 http 伪装在 mihomo 中是独立的 network
		if strings.EqualFold(query.Get("headerType"), "http") {
			result.Network = "http"
			result.HTTPOpts = newHTTPOpts(query)
		}
	case "http", "h2":
		result.Network = "h2"
		result.H2Opts = &h2Opts{
			Host: splitList(query.Get("host")),
			Path: queryOrDefault(query, "path", "/"),
		}
	case "ws", "httpupgrade":
		// mihomo 没有名为 httpupgrade 的 network，通过 ws-opts 的开关启用
		result.Network = "ws"
		options, err := newWSOpts(query, network == "httpupgrade")
		if err != nil {
			return err
		}
		result.WSOpts = options
	case "grpc":
		result.Network = "grpc"
		result.GRPCOpts = &grpcOpts{ServiceName: query.Get("serviceName")}
	case "xhttp":
		if query.Get("extra") != "" {
			return errors.New("暂不支持 xhttp 的 extra 参数")
		}
		result.Network = "xhttp"
		result.XHTTPOpts = &xhttpOpts{
			Path: queryOrDefault(query, "path", "/"),
			Host: query.Get("host"),
			Mode: query.Get("mode"),
		}
	default:
		return fmt.Errorf("不支持的传输方式: %q", network)
	}

	return nil
}

func newHTTPOpts(query url.Values) *httpOpts {
	options := &httpOpts{
		Method: query.Get("method"),
		Path:   splitList(queryOrDefault(query, "path", "/")),
	}
	if hosts := splitList(query.Get("host")); hosts != nil {
		options.Headers = map[string][]string{"Host": hosts}
	}

	return options
}

func newWSOpts(query url.Values, httpUpgrade bool) (*wsOpts, error) {
	options := &wsOpts{
		Path:        queryOrDefault(query, "path", "/"),
		HTTPUpgrade: httpUpgrade,
	}
	if host := query.Get("host"); host != "" {
		options.Headers = map[string]string{"Host": host}
	}

	// 早期数据: httpupgrade 只需打开快速握手，ws 需要指定长度和请求头
	if earlyData := query.Get("ed"); earlyData != "" {
		size, err := strconv.Atoi(earlyData)
		if err != nil {
			return nil, fmt.Errorf("ed 参数无效: %q", earlyData)
		}
		if httpUpgrade {
			options.HTTPUpgradeFastOpen = true
		} else {
			options.MaxEarlyData = size
			options.EarlyDataHeaderName = "Sec-WebSocket-Protocol"
		}
	}
	if header := query.Get("eh"); header != "" {
		options.EarlyDataHeaderName = header
	}

	return options, nil
}

func queryOrDefault(query url.Values, key string, fallback string) string {
	if value := query.Get(key); value != "" {
		return value
	}

	return fallback
}

// queryBool 兼容 true 和 1 两种写法，无法识别的取值视为 false
func queryBool(query url.Values, key string) bool {
	value, _ := strconv.ParseBool(query.Get(key))

	return value
}

func splitList(value string) []string {
	if value == "" {
		return nil
	}

	return strings.Split(value, ",")
}
