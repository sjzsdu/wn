package helper

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/sjzsdu/wn/lang"
	"golang.org/x/crypto/ssh/terminal"
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
			fmt.Print("\n") // 清除当前行
			done <- false   // 发送 false 表示动画已清理完成
			return
		case <-ticker.C:
			fmt.Printf("\r%s %s... ", spinChars[i], lang.T("Thinking"))
			i = (i + 1) % len(spinChars)
		}
	}
}

func ReadFromTerminal(prompt string) (string, error) {
	rlConfig := &readline.Config{
		Prompt:                 prompt,
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		Stdin:                  os.Stdin,
		DisableAutoSaveHistory: true,
		UniqueEditLine:         true,
		FuncGetWidth:           func() int { return readline.GetScreenWidth() },
		FuncIsTerminal: func() bool {
			return readline.IsTerminal(int(os.Stdin.Fd()))
		},
	}

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

		// 检查是否为 Ctrl+Enter
		if strings.Contains(line, "\x0a") {
			inMultiline = true
			buffer.WriteString(line + "\n")
			// 打印当前行并保持提示符
			fmt.Println(line)
			continue
		}

		if inMultiline && strings.TrimSpace(line) == "" {
			return strings.TrimSpace(buffer.String()), nil
		}

		if !inMultiline {
			return strings.TrimSpace(line), nil
		}

		// 多行模式下打印当前行
		fmt.Println(line)
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

// readPipeContent 读取管道输入内容并返回处理后的字符串
func ReadPipeContent() (string, error) {
	// 保存原始的标准输入
	originalStdin := os.Stdin
	defer func() {
		os.Stdin = originalStdin
	}()

	// 读取管道内容
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// 重新设置标准输入为终端
	tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		return "", fmt.Errorf("failed to reopen terminal: %w", err)
	}
	os.Stdin = tty

	// 重新初始化终端
	if err := InitTerminal(); err != nil {
		return "", fmt.Errorf("failed to initialize terminal: %w", err)
	}

	// 确保终端已经准备好
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("failed to initialize terminal")
	}

	return StripAnsiCodes(string(content)), nil
}
