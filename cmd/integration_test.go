//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"strings"
	"testing"

	yt_transcript_api "github.com/dogslee/youtube_transcript_api"
)

// 使用一个公开的、通常有字幕的 YouTube 视频进行测试
// "jNQXAC9IVRw" 是 YouTube 的第一个视频 "Me at the zoo"，通常有字幕
const testVideoID = "jNQXAC9IVRw"

// TestIntegration_FetchTranscript 集成测试：获取字幕
func TestIntegration_FetchTranscript(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api, err := yt_transcript_api.NewYouTubeTranscriptApi(nil)
	if err != nil {
		t.Fatalf("Failed to create API: %v", err)
	}

	transcript, err := api.Fetch(testVideoID, []string{"en"}, false)
	if err != nil {
		t.Fatalf("Failed to fetch transcript: %v", err)
	}

	if transcript == nil {
		t.Fatal("Transcript should not be nil")
	}

	if transcript.VideoID != testVideoID {
		t.Errorf("Expected video ID %s, got %s", testVideoID, transcript.VideoID)
	}

	if len(transcript.Snippets) == 0 {
		t.Error("Transcript should have at least one snippet")
	}

	// 验证字幕片段的基本结构
	for i, snippet := range transcript.Snippets {
		if snippet.Text == "" {
			t.Errorf("Snippet %d should have text", i)
		}
		if snippet.Start < 0 {
			t.Errorf("Snippet %d start time should be >= 0, got %f", i, snippet.Start)
		}
		if snippet.Duration <= 0 {
			t.Errorf("Snippet %d duration should be > 0, got %f", i, snippet.Duration)
		}
	}

	t.Logf("Successfully fetched transcript with %d snippets", len(transcript.Snippets))
}

// TestIntegration_ListTranscripts 集成测试：列出可用字幕
func TestIntegration_ListTranscripts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api, err := yt_transcript_api.NewYouTubeTranscriptApi(nil)
	if err != nil {
		t.Fatalf("Failed to create API: %v", err)
	}

	transcriptList, err := api.List(testVideoID)
	if err != nil {
		t.Fatalf("Failed to list transcripts: %v", err)
	}

	if transcriptList == nil {
		t.Fatal("TranscriptList should not be nil")
	}

	if transcriptList.VideoID != testVideoID {
		t.Errorf("Expected video ID %s, got %s", testVideoID, transcriptList.VideoID)
	}

	// 验证列表字符串表示
	listStr := transcriptList.String()
	if listStr == "" {
		t.Error("TranscriptList string representation should not be empty")
	}

	if !strings.Contains(listStr, testVideoID) {
		t.Error("TranscriptList string should contain video ID")
	}

	t.Logf("Successfully listed transcripts:\n%s", listStr)
}

// TestIntegration_FindTranscript 集成测试：查找指定语言的字幕
func TestIntegration_FindTranscript(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api, err := yt_transcript_api.NewYouTubeTranscriptApi(nil)
	if err != nil {
		t.Fatalf("Failed to create API: %v", err)
	}

	transcriptList, err := api.List(testVideoID)
	if err != nil {
		t.Fatalf("Failed to list transcripts: %v", err)
	}

	// 尝试查找英语字幕
	transcript, err := transcriptList.FindTranscript([]string{"en"})
	if err != nil {
		t.Logf("English transcript not found, trying other languages: %v", err)
		// 如果英语不可用，尝试其他常见语言
		transcript, err = transcriptList.FindTranscript([]string{"en", "zh", "es", "fr"})
		if err != nil {
			t.Fatalf("Failed to find any transcript: %v", err)
		}
	}

	if transcript == nil {
		t.Fatal("Transcript should not be nil")
	}

	t.Logf("Found transcript: %s (%s)", transcript.Language, transcript.LanguageCode)
}

// TestIntegration_Formatters 集成测试：测试各种格式化器
func TestIntegration_Formatters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api, err := yt_transcript_api.NewYouTubeTranscriptApi(nil)
	if err != nil {
		t.Fatalf("Failed to create API: %v", err)
	}

	transcript, err := api.Fetch(testVideoID, []string{"en"}, false)
	if err != nil {
		t.Fatalf("Failed to fetch transcript: %v", err)
	}

	formatterLoader := yt_transcript_api.NewFormatterLoader()

	formats := []string{"json", "pretty", "text", "srt", "webvtt"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			formatter, err := formatterLoader.Load(format)
			if err != nil {
				t.Fatalf("Failed to load formatter %s: %v", format, err)
			}

			output, err := formatter.FormatTranscript(transcript)
			if err != nil {
				t.Fatalf("Failed to format transcript with %s: %v", format, err)
			}

			if output == "" {
				t.Errorf("Formatter %s should produce non-empty output", format)
			}

			// 验证 JSON 格式
			if format == "json" {
				var data []map[string]interface{}
				if err := json.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("JSON formatter output should be valid JSON: %v", err)
				}
			}

			t.Logf("Formatter %s produced %d bytes of output", format, len(output))
		})
	}
}

// TestIntegration_CLI 集成测试：测试命令行工具
func TestIntegration_CLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := yt_transcript_api.CLIConfig{
		VideoIDs:  []string{testVideoID},
		Languages: []string{"en"},
		Format:    "text",
	}

	cli := yt_transcript_api.NewYouTubeTranscriptCLI(config)
	output, err := cli.Run()

	if err != nil {
		t.Fatalf("CLI run failed: %v", err)
	}

	if output == "" {
		t.Error("CLI should produce output")
	}

	// 验证输出包含一些文本内容
	if len(output) < 10 {
		t.Errorf("Output seems too short: %d bytes", len(output))
	}

	t.Logf("CLI produced %d bytes of output", len(output))
}

// TestIntegration_ListTranscriptsCLI 集成测试：测试列出字幕的 CLI 功能
func TestIntegration_ListTranscriptsCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := yt_transcript_api.CLIConfig{
		VideoIDs:        []string{testVideoID},
		ListTranscripts: true,
	}

	cli := yt_transcript_api.NewYouTubeTranscriptCLI(config)
	output, err := cli.Run()

	if err != nil {
		t.Fatalf("CLI run failed: %v", err)
	}

	if output == "" {
		t.Error("CLI should produce output when listing transcripts")
	}

	// 验证输出包含视频 ID
	if !strings.Contains(output, testVideoID) {
		t.Error("Output should contain video ID")
	}

	t.Logf("List transcripts CLI output:\n%s", output)
}

// TestIntegration_InvalidVideoID 集成测试：测试无效视频 ID 的错误处理
func TestIntegration_InvalidVideoID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api, err := yt_transcript_api.NewYouTubeTranscriptApi(nil)
	if err != nil {
		t.Fatalf("Failed to create API: %v", err)
	}

	// 使用一个明显无效的视频 ID
	_, err = api.Fetch("invalid_video_id_12345", []string{"en"}, false)
	if err == nil {
		t.Error("Expected error for invalid video ID")
	}

	// 验证错误类型
	if _, ok := err.(*yt_transcript_api.InvalidVideoId); !ok {
		if _, ok := err.(*yt_transcript_api.VideoUnavailable); !ok {
			if _, ok := err.(*yt_transcript_api.CouldNotRetrieveTranscript); !ok {
				t.Logf("Got unexpected error type: %T, error: %v", err, err)
			}
		}
	}

	t.Logf("Correctly handled invalid video ID: %v", err)
}
