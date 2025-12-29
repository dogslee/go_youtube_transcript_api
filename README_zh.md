# YouTube Transcript API - Go Implementation

[![Go Reference](https://pkg.go.dev/badge/github.com/dogslee/youtube_transcript_api.svg)](https://pkg.go.dev/github.com/dogslee/youtube_transcript_api)
[![Go Report Card](https://goreportcard.com/badge/github.com/dogslee/youtube_transcript_api)](https://goreportcard.com/report/github.com/dogslee/youtube_transcript_api)

这是 YouTube Transcript API 的 Go 语言实现，从 [jdepoix/youtube-transcript-api](https://github.com/jdepoix/youtube-transcript-api) (Python 版本) 转译而来。

本项目保持了与原 Python 版本相同的 API 设计和功能特性，使用 Go 语言重新实现，提供了更好的性能和类型安全。

**文档：** [pkg.go.dev/github.com/dogslee/youtube_transcript_api](https://pkg.go.dev/github.com/dogslee/youtube_transcript_api)

## 功能特性

- 获取 YouTube 视频的可用字幕列表
- 下载指定语言的字幕内容
- 支持手动创建和自动生成的字幕
- 支持字幕翻译
- 多种输出格式（JSON、SRT、WebVTT、纯文本）
- 代理配置支持（通用代理和 Webshare 代理）
- 命令行工具支持批量处理

## 系统要求

- **Go 版本**: 1.19.0 或更高版本

详细兼容性说明请参考 [COMPATIBILITY.md](COMPATIBILITY.md)

## 安装

### 作为库使用

```bash
go get github.com/dogslee/youtube_transcript_api
```

### 安装命令行工具

使用 `go install` 安装命令行工具：

```bash
go install github.com/dogslee/youtube_transcript_api/cmd@latest
```

**说明：**
- 安装路径：`github.com/dogslee/youtube_transcript_api/cmd`
- 安装后的可执行文件名：`cmd`（位于 `$GOPATH/bin` 或 `$HOME/go/bin` 目录）
- 如需自定义可执行文件名，可以使用别名或创建符号链接

安装完成后，确保 `$GOPATH/bin` 或 `$HOME/go/bin` 在您的 `PATH` 环境变量中，然后就可以直接使用：

```bash
cmd dQw4w9WgXcQ
```

**卸载方式：**
如需卸载命令行工具，直接删除对应的二进制文件即可：

```bash
# 删除安装的二进制文件
rm $(go env GOPATH)/bin/cmd
# 或者如果设置了 GOBIN
rm $(go env GOBIN)/cmd
# 或者默认位置
rm ~/go/bin/cmd
```

## 使用示例

### 作为库使用

```go
package main

import (
    "fmt"
    yt "github.com/dogslee/youtube_transcript_api"
)

func main() {
    // 创建 API 实例
    api, err := yt.NewYouTubeTranscriptApi(nil)
    if err != nil {
        panic(err)
    }
    
    // 获取字幕
    transcript, err := api.Fetch("video_id", []string{"en"}, false)
    if err != nil {
        panic(err)
    }
    
    // 打印字幕
    for _, snippet := range transcript.Snippets {
        fmt.Printf("[%.2f] %s\n", snippet.Start, snippet.Text)
    }
}
```

### 获取可用字幕列表

```go
api, _ := yt.NewYouTubeTranscriptApi(nil)
transcriptList, err := api.List("video_id")
if err != nil {
    panic(err)
}

// 查找指定语言的字幕
transcript, err := transcriptList.FindTranscript([]string{"en", "zh"})
if err != nil {
    panic(err)
}

// 获取字幕内容
fetched, err := transcript.Fetch(false)
if err != nil {
    panic(err)
}
```

### 使用代理

```go
// 通用代理
proxyConfig, _ := yt.NewGenericProxyConfig("http://proxy.example.com:8080", "")

// Webshare 代理
proxyConfig, _ := yt.NewWebshareProxyConfig(
    "username",
    "password",
    nil, // filterIPLocations
    10,  // retriesWhenBlocked
    "",  // domainName (使用默认值)
    0,   // proxyPort (使用默认值)
)

api, _ := yt.NewYouTubeTranscriptApi(proxyConfig)
```

### 格式化输出

```go
formatterLoader := yt.NewFormatterLoader()

// JSON 格式
jsonFormatter, _ := formatterLoader.Load("json")
jsonOutput, _ := jsonFormatter.FormatTranscript(transcript)

// SRT 格式
srtFormatter, _ := formatterLoader.Load("srt")
srtOutput, _ := srtFormatter.FormatTranscript(transcript)

// WebVTT 格式
webvttFormatter, _ := formatterLoader.Load("webvtt")
webvttOutput, _ := webvttFormatter.FormatTranscript(transcript)

// 纯文本格式
textFormatter, _ := formatterLoader.Load("text")
textOutput, _ := textFormatter.FormatTranscript(transcript)
```

## 命令行工具

### 安装方式

**方式一：使用 go install（推荐）**

```bash
go install github.com/dogslee/youtube_transcript_api/cmd@latest
```

安装路径说明：
- 模块路径：`github.com/dogslee/youtube_transcript_api/cmd`
- 安装后的可执行文件名：`cmd`
- 安装位置：`$GOPATH/bin/cmd` 或 `$HOME/go/bin/cmd`

**方式二：手动编译**

```bash
cd cmd
go build -o youtube-transcript-api
```

编译后的可执行文件在当前目录，可以移动到系统 PATH 目录或直接使用。

### 使用示例

**注意：** 如果使用 `go install` 安装，可执行文件名为 `cmd`；如果手动编译并指定了名称，则使用编译后的文件名。

```bash
# 获取字幕（使用 go install 安装后）
cmd dQw4w9WgXcQ

# 或者（手动编译后，使用自定义名称）
youtube-transcript-api dQw4w9WgXcQ

# 列出可用字幕
cmd --list-transcripts dQw4w9WgXcQ

# 指定语言
cmd --languages "en zh" dQw4w9WgXcQ

# 指定输出格式
cmd --format json dQw4w9WgXcQ

# 翻译字幕
cmd --translate zh dQw4w9WgXcQ

# 使用代理
cmd --http-proxy "http://proxy.example.com:8080" dQw4w9WgXcQ
```

## API 文档

### YouTubeTranscriptApi

主要的 API 接口。

#### NewYouTubeTranscriptApi(proxyConfig ProxyConfig) (*YouTubeTranscriptApi, error)

创建新的 API 实例。

**参数：**
- `proxyConfig`: 可选的代理配置

**返回：**
- `*YouTubeTranscriptApi`: API 实例
- `error`: 错误信息

#### Fetch(videoID string, languages []string, preserveFormatting bool) (*FetchedTranscript, error)

获取单个视频的字幕。

**参数：**
- `videoID`: 视频 ID（不是完整 URL）
- `languages`: 语言代码列表（按优先级排序）
- `preserveFormatting`: 是否保留 HTML 格式标签

**返回：**
- `*FetchedTranscript`: 获取的字幕
- `error`: 错误信息

#### List(videoID string) (*TranscriptList, error)

获取视频的可用字幕列表。

**参数：**
- `videoID`: 视频 ID

**返回：**
- `*TranscriptList`: 字幕列表
- `error`: 错误信息

### TranscriptList

字幕列表对象。

#### FindTranscript(languageCodes []string) (*Transcript, error)

查找字幕（优先手动创建）。

#### FindManuallyCreatedTranscript(languageCodes []string) (*Transcript, error)

仅查找手动创建的字幕。

#### FindGeneratedTranscript(languageCodes []string) (*Transcript, error)

仅查找自动生成的字幕。

### Transcript

字幕对象。

#### Fetch(preserveFormatting bool) (*FetchedTranscript, error)

获取实际字幕内容。

#### Translate(languageCode string) (*Transcript, error)

翻译到指定语言。

### Formatter

格式化器接口。

#### FormatTranscript(transcript *FetchedTranscript) (string, error)

格式化单个字幕。

#### FormatTranscripts(transcripts []*FetchedTranscript) (string, error)

格式化多个字幕。

## 错误处理

所有错误都实现了 `error` 接口。主要错误类型包括：

- `CouldNotRetrieveTranscript`: 无法获取字幕的基类
- `VideoUnavailable`: 视频不可用
- `VideoUnplayable`: 视频无法播放
- `TranscriptsDisabled`: 字幕已禁用
- `NoTranscriptFound`: 未找到字幕
- `RequestBlocked`: 请求被阻止（IP 封禁）
- `AgeRestricted`: 年龄限制视频
- 等等...

## 注意事项

1. **线程安全**：`YouTubeTranscriptApi` 不是线程安全的，在多线程环境中，每个线程需要创建独立的实例。

2. **IP 封禁**：YouTube 可能会封禁频繁请求的 IP。建议使用代理或轮换 IP。

3. **Cookie 认证**：目前不支持 Cookie 认证，因此无法获取年龄限制视频的字幕。

4. **API 变化**：YouTube 可能会更改其 API 结构，这可能导致某些功能失效。

## 致谢

本项目是从 [jdepoix/youtube-transcript-api](https://github.com/jdepoix/youtube-transcript-api) 转译而来的 Go 语言实现。

感谢原项目作者 [@jdepoix](https://github.com/jdepoix) 及其贡献者们创建并维护了优秀的 Python 版本，为本项目提供了重要的参考和设计基础。

原项目采用 MIT 许可证，本项目同样采用 MIT 许可证，以保持许可证的一致性。

## 许可证

本项目采用 [MIT 许可证](LICENSE) 进行许可。

Copyright (c) 2025 dogslee

MIT 许可证是一个宽松的开源许可证，允许您自由使用、修改和分发代码，只需保留版权声明和许可证文本即可。

