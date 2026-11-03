package main

import (
	"bytes"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

// convertLinks 全部链接转换成功后才返回结果，避免留下只有一部分节点的配置
func convertLinks(links []string) (string, error) {
	proxies := make([]*proxy, 0, len(links))
	nameCount := make(map[string]int, len(links))
	for index, link := range links {
		item, err := parseLink(link)
		if err != nil {
			return "", fmt.Errorf("第 %d 条链接: %w", index+1, err)
		}

		// mihomo 不允许节点重名，重复的名称追加序号
		nameCount[item.Name]++
		if count := nameCount[item.Name]; count > 1 {
			item.Name = fmt.Sprintf("%s-%d", item.Name, count)
		}

		proxies = append(proxies, item)
	}

	return encodeProxies(proxies)
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
