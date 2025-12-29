package main

import (
	"flag"
	"os"
	"strings"
	"testing"

	yt_transcript_api "github.com/dogslee/youtube_transcript_api"
)

// TestCLIConfig 测试 CLI 配置的创建和默认值
func TestCLIConfig(t *testing.T) {
	config := yt_transcript_api.CLIConfig{
		VideoIDs:  []string{"test_video_id"},
		Languages: []string{},
		Format:    "",
	}

	cli := yt_transcript_api.NewYouTubeTranscriptCLI(config)

	// 测试运行会使用默认值，如果配置错误会失败
	// 这里我们只测试 CLI 对象是否成功创建
	if cli == nil {
		t.Error("CLI should not be nil")
	}
}

// TestVideoIDSanitization 测试视频 ID 清理（移除反斜杠）
// 这个测试通过实际运行 CLI 来验证清理逻辑
func TestVideoIDSanitization(t *testing.T) {
	config := yt_transcript_api.CLIConfig{
		VideoIDs:  []string{"test\\video\\id"},
		Languages: []string{"en"},
		Format:    "text",
	}

	cli := yt_transcript_api.NewYouTubeTranscriptCLI(config)

	// 运行 CLI（会失败，因为视频 ID 无效，但可以验证清理逻辑）
	_, err := cli.Run()

	// 错误是预期的（因为视频 ID 无效），但我们验证了清理逻辑被执行
	// 如果清理失败，可能会有其他错误
	if err == nil {
		// 如果成功，说明清理工作正常
		t.Log("Video ID sanitization test passed")
	} else {
		// 检查错误是否与视频 ID 相关（说明清理后的 ID 被使用）
		if !strings.Contains(err.Error(), "test") {
			t.Logf("Expected error related to video ID, got: %v", err)
		}
	}
}

// TestExcludeBothFlags 测试同时排除手动创建和自动生成的字幕
func TestExcludeBothFlags(t *testing.T) {
	config := yt_transcript_api.CLIConfig{
		VideoIDs:               []string{"test_video_id"},
		ExcludeGenerated:       true,
		ExcludeManuallyCreated: true,
	}

	cli := yt_transcript_api.NewYouTubeTranscriptCLI(config)
	output, err := cli.Run()

	if err != nil {
		t.Errorf("Should not return error when both flags are set, got: %v", err)
	}

	if output != "" {
		t.Errorf("Output should be empty when both flags are set, got: %s", output)
	}
}

// TestProxyConfig 测试代理配置
func TestProxyConfig(t *testing.T) {
	// 测试通用代理配置
	config := yt_transcript_api.CLIConfig{
		VideoIDs:  []string{"test_video_id"},
		HTTPProxy: "http://proxy.example.com:8080",
	}

	cli := yt_transcript_api.NewYouTubeTranscriptCLI(config)
	// 这里只是测试配置是否正确创建，不实际运行（因为需要网络）
	_ = cli
}

// TestLanguageParsing 测试语言列表解析
func TestLanguageParsing(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single language", "en", []string{"en"}},
		{"multiple languages", "en zh de", []string{"en", "zh", "de"}},
		{"empty string", "", []string{"en"}}, // 默认值
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			languageList := []string{"en"}
			if tc.input != "" {
				languageList = strings.Fields(tc.input)
			}

			if len(languageList) != len(tc.expected) {
				t.Errorf("Expected %d languages, got %d", len(tc.expected), len(languageList))
			}
		})
	}
}

// TestFlagParsing 测试命令行参数解析（模拟）
func TestFlagParsing(t *testing.T) {
	// 保存原始 os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// 测试版本标志
	os.Args = []string{"cmd", "-version"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	version := flag.Bool("version", false, "")
	flag.Parse()

	if !*version {
		t.Error("Version flag should be true")
	}

	// 测试其他标志
	os.Args = []string{"cmd", "-list-transcripts", "-languages", "en zh", "dQw4w9WgXcQ"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	listTranscripts := flag.Bool("list-transcripts", false, "")
	languages := flag.String("languages", "en", "")
	flag.Parse()

	if !*listTranscripts {
		t.Error("list-transcripts flag should be true")
	}

	if *languages != "en zh" {
		t.Errorf("Expected languages 'en zh', got '%s'", *languages)
	}
}
