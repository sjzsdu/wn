package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// PrintWithLabel 带标签的打印，方便调试时识别输出内容
func PrintWithLabel(label string, v interface{}) {
	fmt.Printf("[%s]: ", label)
	if v == nil {
		fmt.Println("nil")
		return
	}
	Print(v)
}

func Print(v interface{}) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Map, reflect.Slice, reflect.Struct, reflect.Ptr:
		formatted, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("格式化输出失败: %v\n", err)
			return
		}
		fmt.Println(string(formatted))
	default:
		fmt.Println(v)
	}
}

// Printf 支持格式化字符串的打印
func Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// Println 换行打印
func Println(v ...interface{}) {
	fmt.Println(v...)
}
