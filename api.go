// Package youtube_transcript_api provides a Go implementation of the YouTube Transcript API.
// It allows you to retrieve transcripts/subtitles for YouTube videos, including automatically
// generated subtitles, without requiring an API key or headless browser.
//
// This package is a Go port of the Python youtube-transcript-api library
// (https://github.com/jdepoix/youtube-transcript-api).
//
// Example usage:
//
//	api, err := youtube_transcript_api.NewYouTubeTranscriptApi(nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	transcript, err := api.Fetch("video_id", []string{"en"}, false)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, snippet := range transcript.Snippets {
//		fmt.Printf("[%.2f] %s\n", snippet.Start, snippet.Text)
//	}
package youtube_transcript_api

// YouTubeTranscriptApi 主要的 API 接口
type YouTubeTranscriptApi struct {
	fetcher *TranscriptListFetcher
}

// NewYouTubeTranscriptApi 创建新的 YouTubeTranscriptApi 实例
// 注意：由于 HTTPClient 不是线程安全的，在多线程环境中，每个线程需要创建独立的实例
func NewYouTubeTranscriptApi(proxyConfig ProxyConfig) (*YouTubeTranscriptApi, error) {
	httpClient, err := NewHTTPClient()
	if err != nil {
		return nil, err
	}

	// 设置默认请求头
	httpClient.Headers["Accept-Language"] = "en-US"

	// 设置代理
	if proxyConfig != nil {
		if err := SetupHTTPClientProxy(httpClient, proxyConfig); err != nil {
			return nil, err
		}
	}

	fetcher := NewTranscriptListFetcher(httpClient, proxyConfig)

	return &YouTubeTranscriptApi{
		fetcher: fetcher,
	}, nil
}

// Fetch 获取单个视频的字幕
// 这是调用 list().find_transcript(languages).fetch(preserve_formatting) 的快捷方式
func (api *YouTubeTranscriptApi) Fetch(videoID string, languages []string, preserveFormatting bool) (*FetchedTranscript, error) {
	if len(languages) == 0 {
		languages = []string{"en"}
	}

	transcriptList, err := api.List(videoID)
	if err != nil {
		return nil, err
	}

	transcript, err := transcriptList.FindTranscript(languages)
	if err != nil {
		return nil, err
	}

	return transcript.Fetch(preserveFormatting)
}

// List 获取视频的可用字幕列表
func (api *YouTubeTranscriptApi) List(videoID string) (*TranscriptList, error) {
	return api.fetcher.Fetch(videoID)
}
