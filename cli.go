package youtube_transcript_api

import (
	"strings"
)

// CLIConfig 命令行配置
type CLIConfig struct {
	VideoIDs               []string
	ListTranscripts        bool
	Languages              []string
	ExcludeGenerated       bool
	ExcludeManuallyCreated bool
	Format                 string
	Translate              string
	WebshareProxyUsername  string
	WebshareProxyPassword  string
	HTTPProxy              string
	HTTPSProxy             string
}

// YouTubeTranscriptCLI 命令行工具
type YouTubeTranscriptCLI struct {
	config CLIConfig
}

// NewYouTubeTranscriptCLI 创建新的命令行工具实例
func NewYouTubeTranscriptCLI(config CLIConfig) *YouTubeTranscriptCLI {
	// 清理视频 ID（移除反斜杠）
	for i, videoID := range config.VideoIDs {
		config.VideoIDs[i] = strings.ReplaceAll(videoID, "\\", "")
	}

	// 默认语言
	if len(config.Languages) == 0 {
		config.Languages = []string{"en"}
	}

	// 默认格式
	if config.Format == "" {
		config.Format = "pretty"
	}

	return &YouTubeTranscriptCLI{
		config: config,
	}
}

// Run 运行命令行工具
func (cli *YouTubeTranscriptCLI) Run() (string, error) {
	if cli.config.ExcludeManuallyCreated && cli.config.ExcludeGenerated {
		return "", nil
	}

	// 设置代理配置
	var proxyConfig ProxyConfig
	var err error

	if cli.config.HTTPProxy != "" || cli.config.HTTPSProxy != "" {
		proxyConfig, err = NewGenericProxyConfig(cli.config.HTTPProxy, cli.config.HTTPSProxy)
		if err != nil {
			return "", err
		}
	}

	if cli.config.WebshareProxyUsername != "" || cli.config.WebshareProxyPassword != "" {
		proxyConfig, err = NewWebshareProxyConfig(
			cli.config.WebshareProxyUsername,
			cli.config.WebshareProxyPassword,
			nil, // filterIPLocations
			10,  // retriesWhenBlocked
			"",  // domainName (使用默认值)
			0,   // proxyPort (使用默认值)
		)
		if err != nil {
			return "", err
		}
	}

	// 创建 API 实例
	api, err := NewYouTubeTranscriptApi(proxyConfig)
	if err != nil {
		return "", err
	}

	var transcripts []*FetchedTranscript
	var transcriptLists []*TranscriptList
	var exceptions []error

	// 处理每个视频
	for _, videoID := range cli.config.VideoIDs {
		transcriptList, err := api.List(videoID)
		if err != nil {
			exceptions = append(exceptions, err)
			continue
		}

		if cli.config.ListTranscripts {
			transcriptLists = append(transcriptLists, transcriptList)
		} else {
			transcript, err := cli.fetchTranscript(transcriptList)
			if err != nil {
				exceptions = append(exceptions, err)
				continue
			}
			transcripts = append(transcripts, transcript)
		}
	}

	// 构建输出
	var outputSections []string

	// 添加异常信息
	for _, exception := range exceptions {
		outputSections = append(outputSections, exception.Error())
	}

	// 添加字幕数据
	if cli.config.ListTranscripts {
		for _, transcriptList := range transcriptLists {
			outputSections = append(outputSections, transcriptList.String())
		}
	} else if len(transcripts) > 0 {
		formatterLoader := NewFormatterLoader()
		formatter, err := formatterLoader.Load(cli.config.Format)
		if err != nil {
			return "", err
		}

		formatted, err := formatter.FormatTranscripts(transcripts)
		if err != nil {
			return "", err
		}

		outputSections = append(outputSections, formatted)
	}

	return strings.Join(outputSections, "\n\n"), nil
}

func (cli *YouTubeTranscriptCLI) fetchTranscript(transcriptList *TranscriptList) (*FetchedTranscript, error) {
	var transcript *Transcript
	var err error

	if cli.config.ExcludeManuallyCreated {
		transcript, err = transcriptList.FindGeneratedTranscript(cli.config.Languages)
	} else if cli.config.ExcludeGenerated {
		transcript, err = transcriptList.FindManuallyCreatedTranscript(cli.config.Languages)
	} else {
		transcript, err = transcriptList.FindTranscript(cli.config.Languages)
	}

	if err != nil {
		return nil, err
	}

	// 如果需要翻译
	if cli.config.Translate != "" {
		transcript, err = transcript.Translate(cli.config.Translate)
		if err != nil {
			return nil, err
		}
	}

	return transcript.Fetch(false) // preserveFormatting = false
}
