# v2any

## 项目目标

把 `vless://` 分享链接转换为 mihomo（Clash Meta、Clash Verge 的内核）的 `proxies` 配置项。目前只支持这一种输入和输出，名称中的 `any` 是为以后扩展留的余地，需要时再扩展。

## 结构与技术栈

- 单个 Go 命令行程序，代码都在根目录：`main.go` 处理输入和输出，`vless.go` 负责解析链接和映射字段
- 依赖 `go.yaml.in/yaml/v3`。必须用 YAML 库输出，而不是拼接文本：`short-id` 可能是 `12345678`、`0e12` 这类会被解析成数字的取值，必须加引号
- 用 GoReleaser 发布，配置见 `.goreleaser.yaml` 和 `.github/workflows/release.yml`

## 关键约束

- 输出格式以用户给定的示例为准：键的顺序就是 `proxy` 结构体的字段顺序，列表带两个空格缩进，不输出 `proxies:` 这一行。`_spider-x` 这个键名也来自该示例，mihomo 会忽略它
- 不校验 UUID 和 `short-id` 的格式：用户的示例中两者都不符合标准格式，转换结果仍要原样输出
- `httpupgrade` 必须输出为 `network: ws` 加 `ws-opts.v2ray-http-upgrade: true`。mihomo 内置转换器（`common/convert/v.go`）写的是 `network: httpupgrade`，而它的 VLESS 适配器没有这个分支，不要照搬
- 无法识别的 `security` 和传输方式必须报错，不能退化成明文或 `tcp` 后继续输出
- 仓库中不得出现真实节点的链接、UUID、密钥，文档示例一律使用虚构数据

## 开发约定

- 不写单元测试。修改后用真实链接运行，再用 mihomo 校验：`mihomo -d <临时目录> -f <配置文件> -t`。本机可用 `/Applications/Clash Verge.app/Contents/MacOS/verge-mihomo`
- 工作流中的 Action 必须使用最新的主版本，修改前先核对官方仓库的最新 release
- 推送 `master` 只构建快照，推送 `v*` 标签才发布 Release
