# WN - 文件打包工具

## 概述
WN是一个命令行工具，用于将指定类型的源代码文件打包成PDF、Markdown、XML或纯文本格式。

## 功能特性
- 支持多种文件格式（默认支持py, ts, js, html, less）
- 支持多种输出格式（PDF、Markdown、XML、纯文本）
- 支持文件排除模式
- 支持中英文语言切换

## 安装
1. 确保已安装Go 1.20+
2. 克隆仓库
3. 运行`go build`编译项目

## 使用说明
```bash
wn pack [flags]
```

### 参数
- `-e, --exts`：要包含的文件扩展名（默认：py,ts,js,html,less）
- `-o, --output`：输出文件名（默认：output.pdf）
- `-x, --excludes`：要排除的文件模式
- `-p, --workPath`：工作目录

## 示例
1. 打包所有Python文件为PDF：
```bash
wn pack -e py -o output.pdf
```

2. 打包所有JavaScript文件为Markdown，排除test目录：
```bash
wn pack -e js -o output.md -x "test/*"
```

## 配置
- 设置环境变量`WN_LANG=zh`启用中文界面

## 贡献
欢迎提交PR和issue

## 许可证
MIT
