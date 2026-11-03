.PHONY: site

# 构建 GitHub Pages 站点到 _site，wasm_exec.js 必须与编译 wasm 的 Go 版本一致
site:
	rm -rf _site
	mkdir -p _site
	cp site/* _site/
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" _site/
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o _site/v2any.wasm .
