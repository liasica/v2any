package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

const usage = `将 vless 链接转换为 mihomo（Clash Meta）的 proxies 配置项

用法:
  v2any 'vless://...' ['vless://...' ...]
  pbpaste | v2any
`

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	if err := run(flag.Args()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "v2any: %v\n", err)
		os.Exit(1)
	}
}

// run 全部链接转换成功后才输出，避免留下只有一部分节点的配置
func run(arguments []string) error {
	links, err := readLinks(arguments)
	if err != nil {
		return err
	}

	proxies := make([]*proxy, 0, len(links))
	nameCount := make(map[string]int, len(links))
	for index, link := range links {
		var item *proxy
		item, err = parseLink(link)
		if err != nil {
			return fmt.Errorf("第 %d 条链接: %w", index+1, err)
		}

		// mihomo 不允许节点重名，重复的名称追加序号
		nameCount[item.Name]++
		if count := nameCount[item.Name]; count > 1 {
			item.Name = fmt.Sprintf("%s-%d", item.Name, count)
		}

		proxies = append(proxies, item)
	}

	var output string
	output, err = encodeProxies(proxies)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprint(os.Stdout, output)

	return nil
}

// readLinks 优先读取命令行参数，没有参数时从标准输入读取，链接之间以空白字符分隔
func readLinks(arguments []string) ([]string, error) {
	if len(arguments) > 0 {
		return arguments, nil
	}

	// 标准输入是终端说明没有通过管道传入内容，直接提示用法
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		flag.Usage()
		os.Exit(2)
	}

	var content []byte
	content, err = io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	links := strings.Fields(string(content))
	if len(links) == 0 {
		return nil, errors.New("标准输入中没有链接")
	}

	return links, nil
}

// encodeProxies 输出带两个空格缩进的列表，可以直接粘贴到配置文件的 proxies 下
func encodeProxies(proxies []*proxy) (string, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(proxies); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}

	var output strings.Builder
	for _, line := range strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n") {
		_, _ = fmt.Fprintf(&output, "  %s\n", line)
	}

	return output.String(), nil
}
