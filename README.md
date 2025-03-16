# WN - 多功能命令行工具

## 概述
WN 是一个多功能命令行工具，提供文件打包、代码统计等实用功能，帮助开发者更高效地管理和分析代码。

WN 集成了多种实用工具和AI能力，旨在提升开发者的日常工作效率。它不仅提供了强大的文件打包功能，支持多种输出格式和智能文件处理，还包含了详细的代码统计分析工具，帮助开发者更好地理解和管理代码库。通过集成多个大语言模型（如OpenAI、DeepSeek），WN还提供了智能AI对话功能，配合可自定义的Agent系统，能够针对不同场景提供专业的技术支持和建议。无论是代码管理、项目分析还是技术咨询，WN都能作为开发者的得力助手，提供全方位的支持。

## 安装

### 下载安装 
```bash
curl -sSL https://raw.githubusercontent.com/sjzsdu/wn/refs/heads/master/get-wn.sh | bash
```

### 源码安装
1. 确保已安装Go 1.20+
2. 克隆仓库
3. 运行`go build`编译项目

## 主要功能

### 1. 文件打包 (pack)
将指定类型的源代码文件打包成多种格式的文档。(然后喂给知识库训练)

#### 功能特性
- 支持多种输出格式
  - PDF（优化字体渲染，支持中文显示）
  - Markdown
  - XML
  - 纯文本
- 智能文件处理
  - 支持所有文本文件格式
  - 自动识别可读文本文件
  - 支持文件排除模式
  - 支持.gitignore规则
- Git仓库支持
  - 支持直接克隆并打包Git仓库
- 多语言支持
  - 支持中英文界面切换

#### 使用说明
```bash
wn pack [flags]
```

##### 参数说明
- 全局参数
  - `-p, --workPath`：指定工作目录（默认：当前目录）
- 打包参数
  - `-e, --exts`：要包含的文件扩展名（默认：*，表示所有文件）
  - `-o, --output`：输出文件名（默认：output.xml）
  - `-x, --excludes`：要排除的文件模式
  - `-g, --git-url`：Git仓库URL，直接克隆并打包
  - `-d, --disable-gitignore`：禁用.gitignore规则处理

#### 使用示例
1. 打包所有文件为PDF：
```bash
wn pack -o output.pdf
```

2. 打包指定扩展名的文件为Markdown，排除test目录：
```bash
wn pack -e go,md -o output.md -x "test/*"
```

3. 从Git仓库直接克隆并打包：
```bash
wn pack -g https://github.com/sjzsdu/EventTrader.git -o trader-code.pdf
```

### 2. 代码统计 (static)
统计项目代码的各项指标，帮助开发者了解代码结构和质量。

#### 功能特性
- 代码量统计
  - 总行数统计
  - 代码行数统计
  - 注释行数统计
  - 空行统计
- 文件分析
  - 按语言类型分类统计
  - 文件数量统计
  - 文件大小统计
- 智能识别
  - 自动识别编程语言
  - 支持.gitignore规则

#### 使用说明
```bash
wn static [flags]
```

##### 参数说明
- `-p, --path`：指定统计目录（默认：当前目录）
- `-e, --exts`：指定要统计的文件扩展名
- `-x, --excludes`：要排除的文件模式
- `-d, --detail`：显示详细统计信息

#### 使用示例
1. 统计当前目录：
```bash
wn static
```

2. 统计指定目录下的Go文件：
```bash
wn static -p /path/to/project -e go
```

3. 显示详细统计信息：
```bash
wn static -d
```

## 配置说明

### 全局配置
使用 `wn config` 命令管理全局配置：

```bash
wn config [flags]
```

#### 配置项
- `--lang`：设置界面语言（默认：en）
  - 中文界面：`wn config --lang zh`
  - 英文界面：`wn config --lang en`
- `--default_provider`：设置默认LLM提供商
- `--default_agent`：设置默认Agent
- `--deepseek_apikey`：设置DeepSeek API密钥
- `--deepseek_model`：设置DeepSeek默认模型
- `--openai_apikey`：设置OpenAI API密钥
- `--openai_model`：设置OpenAI默认模型
- `--list`：列出所有当前配置

### 3. AI对话 (ai)
与AI助手进行智能对话，支持多个大语言模型。

#### 功能特性
- 多模型支持
  - OpenAI
  - DeepSeek
- 流式输出
- 支持自定义Agent
- 上下文记忆

#### 使用说明
```bash
wn ai [flags]
```

##### 参数说明
- `-c, --provider`：指定LLM提供商
- `-m, --model`：指定使用的模型
- `-t, --max-tokens`：最大响应token数（默认：2000）
- `-a, --agent`：指定使用的Agent
- `--providers`：列出可用的LLM提供商
- `--models`：列出当前提供商可用的模型

### 4. Agent管理
Agent是预设的AI角色，可以帮助完成特定任务。

#### 使用说明
```bash
wn agent [flags]
```

##### 参数说明
- `--list`：列出所有可用的Agent
- `--create <name>`：创建或更新Agent
- `--delete <name>`：删除指定Agent
- `--show <name>`：显示Agent内容
- `--content <text>`：设置Agent内容
- `--file <path>`：从文件读取Agent内容

## 未来功能规划
- 文件差异比较
- 项目文档生成
- 代码质量检查
- 更多功能持续开发中...

## 贡献
欢迎提交PR和issue，一起完善这个工具。

## 许可证
MIT
