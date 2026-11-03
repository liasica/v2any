package main

import (
	"strings"
	"syscall/js"
)

// main 向网页导出转换函数后保持运行，Go 运行时一旦退出，网页就无法再调用
func main() {
	js.Global().Set("convertLinks", js.FuncOf(func(_ js.Value, arguments []js.Value) any {
		output, err := convertLinks(strings.Fields(arguments[0].String()))
		if err != nil {
			return map[string]any{"error": err.Error()}
		}

		return map[string]any{"output": output}
	}))

	select {}
}
