package youtube_transcript_api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Formatter 格式化器接口
type Formatter interface {
	FormatTranscript(transcript *FetchedTranscript) (string, error)
	FormatTranscripts(transcripts []*FetchedTranscript) (string, error)
}

// JSONFormatter JSON 格式输出
type JSONFormatter struct{}

func (f *JSONFormatter) FormatTranscript(transcript *FetchedTranscript) (string, error) {
	data := transcript.ToRawData()
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (f *JSONFormatter) FormatTranscripts(transcripts []*FetchedTranscript) (string, error) {
	var data []interface{}
	for _, transcript := range transcripts {
		data = append(data, transcript.ToRawData())
	}
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// PrettyPrintFormatter 美化打印格式
type PrettyPrintFormatter struct{}

func (f *PrettyPrintFormatter) FormatTranscript(transcript *FetchedTranscript) (string, error) {
	data := transcript.ToRawData()
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (f *PrettyPrintFormatter) FormatTranscripts(transcripts []*FetchedTranscript) (string, error) {
	var data []interface{}
	for _, transcript := range transcripts {
		data = append(data, transcript.ToRawData())
	}
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// TextFormatter 纯文本格式（无时间戳）
type TextFormatter struct{}

func (f *TextFormatter) FormatTranscript(transcript *FetchedTranscript) (string, error) {
	var lines []string
	for _, snippet := range transcript.Snippets {
		lines = append(lines, snippet.Text)
	}
	return strings.Join(lines, "\n"), nil
}

func (f *TextFormatter) FormatTranscripts(transcripts []*FetchedTranscript) (string, error) {
	var sections []string
	for _, transcript := range transcripts {
		formatted, err := f.FormatTranscript(transcript)
		if err != nil {
			return "", err
		}
		sections = append(sections, formatted)
	}
	return strings.Join(sections, "\n\n\n"), nil
}

// TextBasedFormatter 基于文本的格式化器基类（用于 SRT 和 WebVTT）
type TextBasedFormatter struct {
	*TextFormatter
}

func (f *TextBasedFormatter) secondsToTimestamp(time float64) (hours, mins, secs, ms int) {
	hours = int(time / 3600)
	remainder := time - float64(hours)*3600
	mins = int(remainder / 60)
	remainder = remainder - float64(mins)*60
	secs = int(remainder)
	ms = int((remainder - float64(secs)) * 1000)
	return
}

func (f *TextBasedFormatter) formatTranscript(transcript *FetchedTranscript, formatTimestamp func(int, int, int, int) string, formatHeader func([]string) string, formatHelper func(int, string, *FetchedTranscriptSnippet) string) (string, error) {
	var lines []string
	for i := range transcript.Snippets {
		snippet := &transcript.Snippets[i]
		end := snippet.Start + snippet.Duration

		// 如果下一个片段的开始时间小于当前结束时间，使用下一个片段的开始时间
		if i < len(transcript.Snippets)-1 && transcript.Snippets[i+1].Start < end {
			end = transcript.Snippets[i+1].Start
		}

		h1, m1, s1, ms1 := f.secondsToTimestamp(snippet.Start)
		h2, m2, s2, ms2 := f.secondsToTimestamp(end)

		timeText := fmt.Sprintf("%s --> %s",
			formatTimestamp(h1, m1, s1, ms1),
			formatTimestamp(h2, m2, s2, ms2),
		)

		lines = append(lines, formatHelper(i, timeText, snippet))
	}

	return formatHeader(lines), nil
}

// SRTFormatter SRT 字幕文件格式
type SRTFormatter struct {
	*TextBasedFormatter
}

func NewSRTFormatter() *SRTFormatter {
	return &SRTFormatter{
		TextBasedFormatter: &TextBasedFormatter{
			TextFormatter: &TextFormatter{},
		},
	}
}

func (f *SRTFormatter) formatTimestamp(hours, mins, secs, ms int) string {
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, mins, secs, ms)
}

func (f *SRTFormatter) formatHeader(lines []string) string {
	return strings.Join(lines, "\n\n") + "\n"
}

func (f *SRTFormatter) formatHelper(i int, timeText string, snippet *FetchedTranscriptSnippet) string {
	return fmt.Sprintf("%d\n%s\n%s", i+1, timeText, snippet.Text)
}

func (f *SRTFormatter) FormatTranscript(transcript *FetchedTranscript) (string, error) {
	return f.formatTranscript(transcript, f.formatTimestamp, f.formatHeader, f.formatHelper)
}

func (f *SRTFormatter) FormatTranscripts(transcripts []*FetchedTranscript) (string, error) {
	var sections []string
	for _, transcript := range transcripts {
		formatted, err := f.FormatTranscript(transcript)
		if err != nil {
			return "", err
		}
		sections = append(sections, formatted)
	}
	return strings.Join(sections, "\n\n"), nil
}

// WebVTTFormatter WebVTT 字幕文件格式
type WebVTTFormatter struct {
	*TextBasedFormatter
}

func NewWebVTTFormatter() *WebVTTFormatter {
	return &WebVTTFormatter{
		TextBasedFormatter: &TextBasedFormatter{
			TextFormatter: &TextFormatter{},
		},
	}
}

func (f *WebVTTFormatter) formatTimestamp(hours, mins, secs, ms int) string {
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, mins, secs, ms)
}

func (f *WebVTTFormatter) formatHeader(lines []string) string {
	return "WEBVTT\n\n" + strings.Join(lines, "\n\n") + "\n"
}

func (f *WebVTTFormatter) formatHelper(i int, timeText string, snippet *FetchedTranscriptSnippet) string {
	return fmt.Sprintf("%s\n%s", timeText, snippet.Text)
}

func (f *WebVTTFormatter) FormatTranscript(transcript *FetchedTranscript) (string, error) {
	return f.formatTranscript(transcript, f.formatTimestamp, f.formatHeader, f.formatHelper)
}

func (f *WebVTTFormatter) FormatTranscripts(transcripts []*FetchedTranscript) (string, error) {
	var sections []string
	for _, transcript := range transcripts {
		formatted, err := f.FormatTranscript(transcript)
		if err != nil {
			return "", err
		}
		sections = append(sections, formatted)
	}
	return strings.Join(sections, "\n\n"), nil
}

// FormatterLoader 格式化器加载器
type FormatterLoader struct {
	types map[string]func() Formatter
}

// NewFormatterLoader 创建格式化器加载器
func NewFormatterLoader() *FormatterLoader {
	return &FormatterLoader{
		types: map[string]func() Formatter{
			"json":   func() Formatter { return &JSONFormatter{} },
			"pretty": func() Formatter { return &PrettyPrintFormatter{} },
			"text":   func() Formatter { return &TextFormatter{} },
			"webvtt": func() Formatter { return NewWebVTTFormatter() },
			"srt":    func() Formatter { return NewSRTFormatter() },
		},
	}
}

// Load 加载指定类型的格式化器
func (fl *FormatterLoader) Load(formatterType string) (Formatter, error) {
	if formatterType == "" {
		formatterType = "pretty"
	}

	formatterFactory, ok := fl.types[formatterType]
	if !ok {
		var supportedTypes []string
		for k := range fl.types {
			supportedTypes = append(supportedTypes, k)
		}
		return nil, fmt.Errorf("the format '%s' is not supported. Choose one of the following formats: %s",
			formatterType, strings.Join(supportedTypes, ", "))
	}

	return formatterFactory(), nil
}
