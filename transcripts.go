package youtube_transcript_api

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/beevik/etree"
)

// FetchedTranscriptSnippet 表示一个字幕片段
type FetchedTranscriptSnippet struct {
	Text     string  // 字幕文本内容
	Start    float64 // 字幕在视频中出现的开始时间（秒）
	Duration float64 // 字幕在屏幕上显示的持续时间（秒，注意：不是语音时长，可能存在重叠）
}

// FetchedTranscript 表示一个完整的已获取字幕
type FetchedTranscript struct {
	Title        string // 视频标题
	ThumbnailURL string // 视频封面URL
	Snippets     []FetchedTranscriptSnippet
	VideoID      string // 视频ID
	Language     string // 字幕语言
	LanguageCode string // 字幕语言代码
	IsGenerated  bool   // 是否是自动生成的字幕
}

// ToRawData 转换为原始数据格式（用于 JSON 序列化）
func (ft *FetchedTranscript) ToRawData() []map[string]interface{} {
	result := make([]map[string]interface{}, len(ft.Snippets))
	for i, snippet := range ft.Snippets {
		result[i] = map[string]interface{}{
			"text":     snippet.Text,
			"start":    snippet.Start,
			"duration": snippet.Duration,
		}
	}
	return result
}

// TranslationLanguage 表示可翻译的语言
type TranslationLanguage struct {
	Language     string
	LanguageCode string
}

// Transcript 表示一个可用的字幕资源
type Transcript struct {
	httpClient              *HTTPClient
	VideoID                 string
	url                     string
	Title                   string
	ThumbnailURL            string
	Language                string
	LanguageCode            string
	IsGenerated             bool
	TranslationLanguages    []TranslationLanguage
	translationLanguagesMap map[string]string
}

// NewTranscript 创建新的 Transcript 对象
func NewTranscript(
	httpClient *HTTPClient,
	videoID string,
	title string,
	thumbnailURL string,
	url string,
	language string,
	languageCode string,
	isGenerated bool,
	translationLanguages []TranslationLanguage,
) *Transcript {
	translationMap := make(map[string]string)
	for _, tl := range translationLanguages {
		translationMap[tl.LanguageCode] = tl.Language
	}

	return &Transcript{
		httpClient:              httpClient,
		VideoID:                 videoID,
		Title:                   title,
		ThumbnailURL:            thumbnailURL,
		url:                     url,
		Language:                language,
		LanguageCode:            languageCode,
		IsGenerated:             isGenerated,
		TranslationLanguages:    translationLanguages,
		translationLanguagesMap: translationMap,
	}
}

// IsTranslatable 检查是否可翻译
func (t *Transcript) IsTranslatable() bool {
	return len(t.TranslationLanguages) > 0
}

// Fetch 获取实际字幕内容
func (t *Transcript) Fetch(preserveFormatting bool) (*FetchedTranscript, error) {
	if strings.Contains(t.url, "&exp=xpe") {
		return nil, NewPoTokenRequired(t.VideoID)
	}

	resp, err := t.httpClient.Get(t.url)
	if err != nil {
		return nil, NewYouTubeRequestFailed(t.VideoID, err)
	}
	defer resp.Body.Close()

	if err := raiseHTTPErrors(resp, t.VideoID); err != nil {
		return nil, err
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewYouTubeRequestFailed(t.VideoID, err)
	}
	body := string(bodyBytes)

	parser := NewTranscriptParser(preserveFormatting)
	snippets, err := parser.Parse(body)
	if err != nil {
		return nil, NewYouTubeRequestFailed(t.VideoID, err)
	}

	return &FetchedTranscript{
		Title:        t.Title,
		ThumbnailURL: t.ThumbnailURL,
		Snippets:     snippets,
		VideoID:      t.VideoID,
		Language:     t.Language,
		LanguageCode: t.LanguageCode,
		IsGenerated:  t.IsGenerated,
	}, nil
}

// Translate 翻译到指定语言
func (t *Transcript) Translate(languageCode string) (*Transcript, error) {
	if !t.IsTranslatable() {
		return nil, NewNotTranslatable(t.VideoID)
	}

	translatedLanguage, ok := t.translationLanguagesMap[languageCode]
	if !ok {
		return nil, NewTranslationLanguageNotAvailable(t.VideoID)
	}

	// 构建翻译后的 URL
	translatedURL := fmt.Sprintf("%s&tlang=%s", t.url, languageCode)

	return NewTranscript(
		t.httpClient,
		t.VideoID,
		t.Title,
		t.ThumbnailURL,
		translatedURL,
		translatedLanguage,
		languageCode,
		true,                    // 翻译后的字幕标记为自动生成
		[]TranslationLanguage{}, // 翻译后的字幕不能再翻译
	), nil
}

// String 返回字符串表示
func (t *Transcript) String() string {
	translationDesc := ""
	if t.IsTranslatable() {
		translationDesc = "[TRANSLATABLE]"
	}
	return fmt.Sprintf(`%s ("%s")%s`, t.LanguageCode, t.Language, translationDesc)
}

// TranscriptList 表示某个视频的所有可用字幕列表
type TranscriptList struct {
	VideoID                    string
	manuallyCreatedTranscripts map[string]*Transcript
	generatedTranscripts       map[string]*Transcript
	translationLanguages       []TranslationLanguage
}

// NewTranscriptList 创建新的 TranscriptList
func NewTranscriptList(
	videoID string,
	manuallyCreatedTranscripts map[string]*Transcript,
	generatedTranscripts map[string]*Transcript,
	translationLanguages []TranslationLanguage,
) *TranscriptList {
	return &TranscriptList{
		VideoID:                    videoID,
		manuallyCreatedTranscripts: manuallyCreatedTranscripts,
		generatedTranscripts:       generatedTranscripts,
		translationLanguages:       translationLanguages,
	}
}

// BuildTranscriptList 从 JSON 数据构建 TranscriptList
func BuildTranscriptList(httpClient *HTTPClient, videoID string, videoDetailsJSON map[string]interface{}, captionsJSON map[string]interface{}) (*TranscriptList, error) {
	// 解析翻译语言
	var translationLanguages []TranslationLanguage
	if translationLangs, ok := captionsJSON["translationLanguages"].([]interface{}); ok {
		for _, tl := range translationLangs {
			if tlMap, ok := tl.(map[string]interface{}); ok {
				if langName, ok := tlMap["languageName"].(map[string]interface{}); ok {
					if runs, ok := langName["runs"].([]interface{}); ok && len(runs) > 0 {
						if run, ok := runs[0].(map[string]interface{}); ok {
							if text, ok := run["text"].(string); ok {
								if langCode, ok := tlMap["languageCode"].(string); ok {
									translationLanguages = append(translationLanguages, TranslationLanguage{
										Language:     text,
										LanguageCode: langCode,
									})
								}
							}
						}
					}
				}
			}
		}
	}

	manuallyCreatedTranscripts := make(map[string]*Transcript)
	generatedTranscripts := make(map[string]*Transcript)

	// 解析字幕轨道
	if captionTracks, ok := captionsJSON["captionTracks"].([]interface{}); ok {
		for _, caption := range captionTracks {
			if captionMap, ok := caption.(map[string]interface{}); ok {
				kind, _ := captionMap["kind"].(string)
				isGenerated := kind == "asr"

				var transcriptDict map[string]*Transcript
				if isGenerated {
					transcriptDict = generatedTranscripts
				} else {
					transcriptDict = manuallyCreatedTranscripts
				}

				// 提取语言代码
				languageCode, ok := captionMap["languageCode"].(string)
				if !ok {
					continue
				}

				// 提取语言名称
				var languageName string
				if name, ok := captionMap["name"].(map[string]interface{}); ok {
					if runs, ok := name["runs"].([]interface{}); ok && len(runs) > 0 {
						if run, ok := runs[0].(map[string]interface{}); ok {
							languageName, _ = run["text"].(string)
						}
					}
				}

				// 提取 baseUrl
				baseURL, ok := captionMap["baseUrl"].(string)
				if !ok {
					continue
				}
				// 移除 &fmt=srv3 参数
				baseURL = strings.ReplaceAll(baseURL, "&fmt=srv3", "")

				// 检查是否可翻译
				var translationLangs []TranslationLanguage
				if isTranslatable, ok := captionMap["isTranslatable"].(bool); ok && isTranslatable {
					translationLangs = translationLanguages
				}

				transcriptDict[languageCode] = NewTranscript(
					httpClient,
					videoID,
					videoDetailsJSON["title"].(string),
					fmt.Sprintf(ThumbnailURLTemplate, videoID),
					baseURL,
					languageName,
					languageCode,
					isGenerated,
					translationLangs,
				)
			}
		}
	}

	return NewTranscriptList(
		videoID,
		manuallyCreatedTranscripts,
		generatedTranscripts,
		translationLanguages,
	), nil
}

// FindTranscript 查找字幕（优先手动创建）
func (tl *TranscriptList) FindTranscript(languageCodes []string) (*Transcript, error) {
	transcriptDicts := []map[string]*Transcript{
		tl.manuallyCreatedTranscripts,
		tl.generatedTranscripts,
	}
	return tl.findTranscript(languageCodes, transcriptDicts)
}

// FindManuallyCreatedTranscript 仅查找手动创建的字幕
func (tl *TranscriptList) FindManuallyCreatedTranscript(languageCodes []string) (*Transcript, error) {
	transcriptDicts := []map[string]*Transcript{
		tl.manuallyCreatedTranscripts,
	}
	return tl.findTranscript(languageCodes, transcriptDicts)
}

// FindGeneratedTranscript 仅查找自动生成的字幕
func (tl *TranscriptList) FindGeneratedTranscript(languageCodes []string) (*Transcript, error) {
	transcriptDicts := []map[string]*Transcript{
		tl.generatedTranscripts,
	}
	return tl.findTranscript(languageCodes, transcriptDicts)
}

func (tl *TranscriptList) findTranscript(languageCodes []string, transcriptDicts []map[string]*Transcript) (*Transcript, error) {
	for _, languageCode := range languageCodes {
		for _, transcriptDict := range transcriptDicts {
			if transcript, ok := transcriptDict[languageCode]; ok {
				return transcript, nil
			}
		}
	}
	return nil, NewNoTranscriptFound(tl.VideoID, languageCodes, tl)
}

// String 返回字符串表示
func (tl *TranscriptList) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("For this video (%s) transcripts are available in the following languages:\n\n", tl.VideoID))

	sb.WriteString("(MANUALLY CREATED)\n")
	sb.WriteString(tl.getLanguageDescription(tl.manuallyCreatedTranscripts))
	sb.WriteString("\n\n")

	sb.WriteString("(GENERATED)\n")
	sb.WriteString(tl.getLanguageDescription(tl.generatedTranscripts))
	sb.WriteString("\n\n")

	sb.WriteString("(TRANSLATION LANGUAGES)\n")
	if len(tl.translationLanguages) == 0 {
		sb.WriteString("None")
	} else {
		for _, tl := range tl.translationLanguages {
			sb.WriteString(fmt.Sprintf(" - %s (\"%s\")\n", tl.LanguageCode, tl.Language))
		}
	}

	return sb.String()
}

func (tl *TranscriptList) getLanguageDescription(transcripts map[string]*Transcript) string {
	if len(transcripts) == 0 {
		return "None"
	}
	var sb strings.Builder
	for _, transcript := range transcripts {
		sb.WriteString(fmt.Sprintf(" - %s\n", transcript.String()))
	}
	return sb.String()
}

// PlayabilityStatus 视频可播放性状态
type PlayabilityStatus string

const (
	PlayabilityStatusOK            PlayabilityStatus = "OK"
	PlayabilityStatusError         PlayabilityStatus = "ERROR"
	PlayabilityStatusLoginRequired PlayabilityStatus = "LOGIN_REQUIRED"
)

// PlayabilityFailedReason 视频无法播放的原因
type PlayabilityFailedReason string

const (
	PlayabilityFailedReasonBotDetected      PlayabilityFailedReason = "Sign in to confirm you're not a bot"
	PlayabilityFailedReasonAgeRestricted    PlayabilityFailedReason = "This video may be inappropriate for some users."
	PlayabilityFailedReasonVideoUnavailable PlayabilityFailedReason = "This video is unavailable"
)

// TranscriptListFetcher 字幕列表获取器
type TranscriptListFetcher struct {
	httpClient  *HTTPClient
	proxyConfig ProxyConfig
}

// NewTranscriptListFetcher 创建新的 TranscriptListFetcher
func NewTranscriptListFetcher(httpClient *HTTPClient, proxyConfig ProxyConfig) *TranscriptListFetcher {
	return &TranscriptListFetcher{
		httpClient:  httpClient,
		proxyConfig: proxyConfig,
	}
}

// Fetch 获取视频的字幕列表
func (tlf *TranscriptListFetcher) Fetch(videoID string) (*TranscriptList, error) {
	videoDetailsJSON, captionsJSON, err := tlf.fetchVideoDetailsAndCaptionsJSON(videoID, 0)
	if err != nil {
		return nil, err
	}

	return BuildTranscriptList(tlf.httpClient, videoID, videoDetailsJSON, captionsJSON)
}

func (tlf *TranscriptListFetcher) fetchVideoDetailsAndCaptionsJSON(videoID string, tryNumber int) (map[string]interface{}, map[string]interface{}, error) {
	html, err := tlf.fetchVideoHTML(videoID)
	if err != nil {
		return nil, nil, err
	}

	apiKey, err := tlf.extractInnertubeAPIKey(html, videoID)
	if err != nil {
		return nil, nil, err
	}

	innertubeData, err := tlf.fetchInnertubeData(videoID, apiKey)
	if err != nil {
		return nil, nil, err
	}

	videoDetailsJSON, captionsJSON, err := tlf.extractVideoDetailsAndCaptionsJSON(innertubeData, videoID)
	if err != nil {
		// 检查是否是 RequestBlocked 错误，如果是且配置了代理，则重试
		if requestBlocked, ok := err.(*RequestBlocked); ok {
			retries := 0
			if tlf.proxyConfig != nil {
				retries = tlf.proxyConfig.RetriesWhenBlocked()
			}
			if tryNumber+1 < retries {
				// 等待一小段时间后重试（触发 IP 轮换）
				time.Sleep(time.Second * time.Duration(tryNumber+1))
				return tlf.fetchVideoDetailsAndCaptionsJSON(videoID, tryNumber+1)
			}
			return nil, nil, requestBlocked.WithProxyConfig(tlf.proxyConfig)
		}
		return nil, nil, err
	}

	return videoDetailsJSON, captionsJSON, nil
}

func (tlf *TranscriptListFetcher) extractInnertubeAPIKey(html, videoID string) (string, error) {
	pattern := regexp.MustCompile(`"INNERTUBE_API_KEY":\s*"([a-zA-Z0-9_-]+)"`)
	matches := pattern.FindStringSubmatch(html)
	if len(matches) == 2 {
		return matches[1], nil
	}

	if strings.Contains(html, `class="g-recaptcha"`) {
		return "", NewIpBlocked(videoID)
	}

	return "", NewYouTubeDataUnparsable(videoID)
}

func (tlf *TranscriptListFetcher) extractVideoDetailsAndCaptionsJSON(innertubeData map[string]interface{}, videoID string) (map[string]interface{}, map[string]interface{}, error) {
	// 检查视频可播放性
	if err := tlf.assertPlayability(innertubeData, videoID); err != nil {
		return nil, nil, err
	}

	// 提取视频详情数据
	videoDetailsJSON, ok := innertubeData["videoDetails"].(map[string]interface{})
	if !ok {
		return nil, nil, NewYouTubeDataUnparsable(videoID)
	}

	// 提取字幕数据
	captions, ok := innertubeData["captions"].(map[string]interface{})
	if !ok {
		return nil, nil, NewTranscriptsDisabled(videoID)
	}

	captionsJSON, ok := captions["playerCaptionsTracklistRenderer"].(map[string]interface{})
	if !ok {
		return nil, nil, NewTranscriptsDisabled(videoID)
	}

	if _, ok := captionsJSON["captionTracks"]; !ok {
		return nil, nil, NewTranscriptsDisabled(videoID)
	}

	return videoDetailsJSON, captionsJSON, nil
}

func (tlf *TranscriptListFetcher) assertPlayability(innertubeData map[string]interface{}, videoID string) error {
	playabilityStatusData, ok := innertubeData["playabilityStatus"].(map[string]interface{})
	if !ok {
		return nil // 如果没有 playabilityStatus，假设可以播放
	}

	status, ok := playabilityStatusData["status"].(string)
	if !ok || status == string(PlayabilityStatusOK) {
		return nil
	}

	reason, _ := playabilityStatusData["reason"].(string)

	if status == string(PlayabilityStatusLoginRequired) {
		if reason == string(PlayabilityFailedReasonBotDetected) {
			return NewRequestBlocked(videoID)
		}
		if reason == string(PlayabilityFailedReasonAgeRestricted) {
			return NewAgeRestricted(videoID)
		}
	}

	if status == string(PlayabilityStatusError) && reason == string(PlayabilityFailedReasonVideoUnavailable) {
		if strings.HasPrefix(videoID, "http://") || strings.HasPrefix(videoID, "https://") {
			return NewInvalidVideoId(videoID)
		}
		return NewVideoUnavailable(videoID)
	}

	// 提取子原因
	var subReasons []string
	if errorScreen, ok := playabilityStatusData["errorScreen"].(map[string]interface{}); ok {
		if playerError, ok := errorScreen["playerErrorMessageRenderer"].(map[string]interface{}); ok {
			if subreason, ok := playerError["subreason"].(map[string]interface{}); ok {
				if runs, ok := subreason["runs"].([]interface{}); ok {
					for _, run := range runs {
						if runMap, ok := run.(map[string]interface{}); ok {
							if text, ok := runMap["text"].(string); ok {
								subReasons = append(subReasons, text)
							}
						}
					}
				}
			}
		}
	}

	return NewVideoUnplayable(videoID, reason, subReasons)
}

func (tlf *TranscriptListFetcher) createConsentCookie(html, videoID string) error {
	pattern := regexp.MustCompile(`name="v" value="(.*?)"`)
	matches := pattern.FindStringSubmatch(html)
	if len(matches) != 2 {
		return NewFailedToCreateConsentCookie(videoID)
	}

	// 设置 Cookie
	cookie := &http.Cookie{
		Name:   "CONSENT",
		Value:  "YES+" + matches[1],
		Domain: ".youtube.com",
	}
	tlf.httpClient.Jar.SetCookies(&url.URL{Scheme: "https", Host: "youtube.com"}, []*http.Cookie{cookie})

	return nil
}

func (tlf *TranscriptListFetcher) fetchVideoHTML(videoID string) (string, error) {
	html, err := tlf.fetchHTML(videoID)
	if err != nil {
		return "", err
	}

	if strings.Contains(html, `action="https://consent.youtube.com/s"`) {
		if err := tlf.createConsentCookie(html, videoID); err != nil {
			return "", err
		}
		html, err = tlf.fetchHTML(videoID)
		if err != nil {
			return "", err
		}
		if strings.Contains(html, `action="https://consent.youtube.com/s"`) {
			return "", NewFailedToCreateConsentCookie(videoID)
		}
	}

	return html, nil
}

func (tlf *TranscriptListFetcher) fetchHTML(videoID string) (string, error) {
	url := fmt.Sprintf(WatchURLTemplate, videoID)
	resp, err := tlf.httpClient.Get(url)
	if err != nil {
		return "", NewYouTubeRequestFailed(videoID, err)
	}
	defer resp.Body.Close()

	if err := raiseHTTPErrors(resp, videoID); err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewYouTubeRequestFailed(videoID, err)
	}

	return html.UnescapeString(string(bodyBytes)), nil
}

func (tlf *TranscriptListFetcher) fetchInnertubeData(videoID, apiKey string) (map[string]interface{}, error) {
	url := fmt.Sprintf(InnertubeAPIURLTemplate, apiKey)

	// 构建请求体
	requestBody := map[string]interface{}{
		"context": InnertubeContext["context"],
		"videoId": videoID,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, NewYouTubeRequestFailed(videoID, err)
	}

	resp, err := tlf.httpClient.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, NewYouTubeRequestFailed(videoID, err)
	}
	defer resp.Body.Close()

	if err := raiseHTTPErrors(resp, videoID); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewYouTubeRequestFailed(videoID, err)
	}

	return result, nil
}

// TranscriptParser 字幕解析器
type TranscriptParser struct {
	preserveFormatting bool
	formattingTags     []string
}

// NewTranscriptParser 创建新的字幕解析器
func NewTranscriptParser(preserveFormatting bool) *TranscriptParser {
	return &TranscriptParser{
		preserveFormatting: preserveFormatting,
		formattingTags: []string{
			"strong", "em", "b", "i", "mark", "small", "del", "ins", "sub", "sup",
		},
	}
}

// Parse 解析 XML 字幕数据
func (tp *TranscriptParser) Parse(rawData string) ([]FetchedTranscriptSnippet, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(rawData); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	root := doc.Root()
	if root == nil {
		return nil, fmt.Errorf("empty XML document")
	}

	var snippets []FetchedTranscriptSnippet

	for _, element := range root.ChildElements() {
		if element.Tag != "text" {
			continue
		}

		text := element.Text()
		if text == "" {
			continue
		}

		// 提取属性
		startStr := element.SelectAttrValue("start", "0.0")
		durationStr := element.SelectAttrValue("dur", "0.0")

		var start, duration float64
		fmt.Sscanf(startStr, "%f", &start)
		fmt.Sscanf(durationStr, "%f", &duration)

		// 处理 HTML 标签
		text = html.UnescapeString(text)
		if !tp.preserveFormatting {
			// 移除所有 HTML 标签
			text = tp.removeAllHTMLTags(text)
		} else {
			// 只保留指定的格式标签
			text = tp.removeNonFormattingHTMLTags(text)
		}

		snippets = append(snippets, FetchedTranscriptSnippet{
			Text:     text,
			Start:    start,
			Duration: duration,
		})
	}

	return snippets, nil
}

func (tp *TranscriptParser) removeAllHTMLTags(text string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(text, "")
}

func (tp *TranscriptParser) removeNonFormattingHTMLTags(text string) string {
	// 匹配所有 HTML 标签
	tagRe := regexp.MustCompile(`(?i)</?([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>`)

	// 创建格式化标签的映射以便快速查找
	formattingTagMap := make(map[string]bool)
	for _, tag := range tp.formattingTags {
		formattingTagMap[strings.ToLower(tag)] = true
	}

	// 替换所有非格式化标签
	result := tagRe.ReplaceAllStringFunc(text, func(match string) string {
		// 提取标签名
		matches := tagRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return "" // 如果无法提取标签名，删除该标签
		}

		tagName := strings.ToLower(matches[1])

		// 如果是格式化标签，保留；否则删除
		if formattingTagMap[tagName] {
			return match
		}
		return ""
	})

	return result
}
