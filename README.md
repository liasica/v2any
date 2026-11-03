# v2any

把 `vless://` 分享链接转换为 [mihomo](https://github.com/MetaCubeX/mihomo)（Clash Meta、Clash Verge 的内核）的 `proxies` 配置项。

## 安装

```bash
go install github.com/liasica/v2any@latest
```

也可以在 Releases 页面下载编译好的二进制文件。

## 使用

链接中含有 `&`，作为参数传入时要加单引号。也可以通过管道传入，多条链接以空白字符分隔：

```bash
v2any 'vless://...'
```

```bash
pbpaste | v2any
```

输入：

```text
vless://00000000-0000-4000-8000-000000000000@example.com:443?encryption=none&flow=xtls-rprx-vision&fp=chrome&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&security=reality&sid=0123456789abcdef&sni=www.example.com&spx=%2F&type=tcp#example.com
```

输出（带两个空格缩进，可以直接粘贴到配置文件的 `proxies:` 下）：

```yaml
  - type: vless
    name: example.com:443
    server: example.com
    port: 443
    uuid: 00000000-0000-4000-8000-000000000000
    udp: true
    tls: true
    flow: xtls-rprx-vision
    client-fingerprint: chrome
    skip-cert-verify: false
    xudp: true
    reality-opts:
      public-key: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      short-id: 0123456789abcdef
      _spider-x: /
    network: tcp
    encryption: none
    servername: www.example.com
```

## 转换规则

| 链接中的 `type` | 输出的 `network` | 选项 |
|---|---|---|
| `tcp`、`raw`，或省略 | `tcp` | 无 |
| `tcp` 且 `headerType=http` | `http` | `http-opts` |
| `http`、`h2` | `h2` | `h2-opts` |
| `ws` | `ws` | `ws-opts` |
| `httpupgrade` | `ws` | `ws-opts`，其中 `v2ray-http-upgrade: true` |
| `grpc` | `grpc` | `grpc-opts` |
| `xhttp` | `xhttp` | `xhttp-opts` |

- **节点名称**：取链接的备注（`#` 之后的部分）。备注为空或与服务器地址相同时，使用「地址:端口」。名称重复时追加序号。
- **默认值**：启用 TLS 且未指定 `fp` 时使用 `chrome`；未指定 `encryption` 时使用 `none`；未指定 `packetEncoding` 时使用 `xudp`。
- **报错退出**：不支持的传输方式（如 `kcp`）、不支持的 `security` 取值，以及 xhttp 的 `extra` 参数。多条链接中只要有一条出错，就不会输出任何内容。
- **不会转换**：`ech`、`pqv` 等其他参数会被忽略。

mihomo 要求 `short-id` 是十六进制字符串。本工具不校验它的格式，如果链接中的取值不合法，mihomo 加载配置时会报 `invalid REALITY short ID`。

## 发布

推送 `v*` 标签后，GitHub Actions 会通过 GoReleaser 构建并发布 Release：

```bash
git tag v0.1.0
```

```bash
git push origin v0.1.0
```
