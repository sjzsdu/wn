package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试用的结构体
type TestStruct struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Hobbies []string `json:"hobbies"`
	Info    struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"info"`
}

func TestMapToStruct(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		want    *TestStruct
		wantErr bool
	}{
		{
			name: "正常转换",
			input: map[string]interface{}{
				"name": "张三",
				"age":  25,
				"hobbies": []string{
					"读书",
					"游泳",
				},
				"info": map[string]interface{}{
					"city":    "北京",
					"country": "中国",
				},
			},
			want: &TestStruct{
				Name: "张三",
				Age:  25,
				Hobbies: []string{
					"读书",
					"游泳",
				},
				Info: struct {
					City    string `json:"city"`
					Country string `json:"country"`
				}{
					City:    "北京",
					Country: "中国",
				},
			},
			wantErr: false,
		},
		{
			name: "类型不匹配",
			input: map[string]interface{}{
				"name": "张三",
				"age":  "25", // 错误的类型：字符串而不是整数
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "空输入",
			input:   map[string]interface{}{},
			want:    &TestStruct{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapToStruct[TestStruct](tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}