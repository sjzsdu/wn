package helper

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/sjzsdu/wn/lang"
)

var (
	rl *readline.Instance
)

// showLoadingAnimation 函数也需要优化以支持取消
func ShowLoadingAnimation(done chan bool) {
	spinChars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// 修改清理行的方式，确保不会影响后续输出
			fmt.Print("\r\033[K")
			return
		case <-ticker.C:
			fmt.Printf("\r%-50s", fmt.Sprintf("%s "+lang.T("Thinking")+"... ", spinChars[i]))
			i = (i + 1) % len(spinChars)
		}
	}
}

// ReadFromTerminal 从终端读取输入，支持多行输入
// prompt: 输入提示符
// returns: 输入内容和可能的错误
// ReadFromTerminal 从终端读取输入，支持多行输入
func ReadFromTerminal(prompt string) (string, error) {
	rlConfig := &readline.Config{
		Prompt:                 prompt,
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		Stdin:                  os.Stdin,
		DisableAutoSaveHistory: true,
		// 添加这些配置以更好地处理管道输入
		UniqueEditLine: true,
		FuncGetWidth:   func() int { return readline.GetScreenWidth() },
		FuncIsTerminal: func() bool {
			return readline.IsTerminal(int(os.Stdin.Fd()))
		},
	}

	// 每次都创建新的实例
	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		return "", fmt.Errorf("初始化readline失败: %v", err)
	}
	defer rl.Close()

	var buffer strings.Builder
	inMultiline := false

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				return "", err
			}
			if err == io.EOF {
				return "exit", nil
			}
			return "", fmt.Errorf("读取输入失败: %v", err)
		}

		// 检查是否为 Ctrl+Enter（通常表示为 \x0a）
		if strings.Contains(line, "\x0a") {
			inMultiline = true
			buffer.WriteString(line + "\n")
			continue
		}

		// 如果是多行模式且收到空行，则结束输入
		if inMultiline && strings.TrimSpace(line) == "" {
			return strings.TrimSpace(buffer.String()), nil
		}

		// 单行模式直接返回
		if !inMultiline {
			return strings.TrimSpace(line), nil
		}

		buffer.WriteString(line + "\n")
	}
}

// InitTerminal 初始化终端
func InitTerminal() error {
	if rl != nil {
		rl.Close()
		rl = nil
	}

	rlConfig := &readline.Config{
		Prompt:          "> ",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		// 添加这些配置以更好地处理管道输入
		UniqueEditLine: true,
		FuncIsTerminal: func() bool {
			return readline.IsTerminal(int(os.Stdin.Fd()))
		},
	}

	var err error
	rl, err = readline.NewEx(rlConfig)
	return err
}
