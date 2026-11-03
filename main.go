//go:build !js

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
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

func run(arguments []string) error {
	links, err := readLinks(arguments)
	if err != nil {
		return err
	}

	var output string
	output, err = convertLinks(links)
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
