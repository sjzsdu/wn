# WN - 多功能命令行工具

## 概述
WN 是一个多功能命令行工具，当前主要提供文件打包功能，未来将扩展更多实用功能。

## 当前功能

### 文件打包 (pack)
将指定类型的源代码文件打包成PDF、Markdown、XML或纯文本格式。

#### 功能特性
- 支持多种文件格式（默认支持所有文本文件）
- 支持多种输出格式（PDF、Markdown、XML、纯文本）
- 支持文件排除模式
- 支持.gitignore规则处理
- 支持从Git仓库直接克隆并打包
- 自动识别可读文本文件
- 支持中英文语言切换

#### 使用说明
```bash
wn pack [flags]
```

##### 全局参数
- `-p, --workPath`：指定工作目录（默认：当前目录）

##### 打包参数
- `-e, --exts`：要包含的文件扩展名（默认：*，表示所有文件）
- `-o, --output`：输出文件名（默认：output.xml）
- `-x, --excludes`：要排除的文件模式
- `-g, --git-url`：Git仓库URL，直接克隆并打包
- `-d, --disable-gitignore`：禁用.gitignore规则处理

#### 示例
1. 打包所有文件为PDF：
```bash
wn pack -o output.xml
```

2. 打包指定扩展名的文件为Markdown，排除test目录：
```bash
wn pack -e go,md -o output.md -x "test/*"
```

3. 从Git仓库直接克隆并打包：
```bash
wn pack -g https://github.com/sjzsdu/EventTrader.git -o trader-code.pdf
```

4. 指定工作目录打包：
```bash
wn pack -p /path/to/project -e go -o project-code.pdf
```

## 未来功能规划
- 文件差异比较
- 代码统计
- 项目文档生成
- 代码质量检查

## 安装

### 下载安装 
```
curl -sSL https://github.com/sjzsdu/wn/raw/master/get-wn.sh | bash
```

### 源码安装
1. 确保已安装Go 1.20+
2. 克隆仓库
3. 运行`go build`编译项目

## 配置
- 设置环境变量`WN_LANG=zh`启用中文界面

## 贡献
欢迎提交PR和issue

## 许可证
MIT
