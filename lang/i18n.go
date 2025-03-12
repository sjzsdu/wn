package lang

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	bundle *i18n.Bundle
	loc    *i18n.Localizer
	// LocalePath 用于配置语言文件的路径
	LocalePath = "lang/locales"
)

// Init initializes the i18n system
func Init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 获取环境变量设置的语言
	lang := os.Getenv("WN_LANG")
	if lang == "" {
		return
	}

	// 标准化语言代码
	switch lang {
	case "zh":
		lang = "zh-CN"
	case "cn":
		lang = "zh-CN"
	case "tw":
		lang = "zh-TW"
	}

	// 检查对应的语言文件是否存在
	langFile := filepath.Join(LocalePath, lang+".json")
	if _, err := os.Stat(langFile); err == nil {
		bundle.MustLoadMessageFile(langFile)
		loc = i18n.NewLocalizer(bundle, lang)
	}
}

// T translates a message
func T(msgID string) string {
	// 如果未初始化 localizer，直接返回原始键
	if loc == nil {
		return msgID
	}

	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID: msgID,
	})
	if err != nil {
		// 如果翻译出错（比如键不存在），返回原始键
		return msgID
	}
	return msg
}
