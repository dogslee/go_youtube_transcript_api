package youtube_transcript_api

import (
	"fmt"
	"net/http"
)

// YouTubeTranscriptApiException 是所有异常的基类
type YouTubeTranscriptApiException struct {
	Message string
}

func (e *YouTubeTranscriptApiException) Error() string {
	return e.Message
}

// CookieError Cookie 相关错误
type CookieError struct {
	*YouTubeTranscriptApiException
}

// CookiePathInvalid Cookie 路径无效
type CookiePathInvalid struct {
	*CookieError
	Path string
}

func NewCookiePathInvalid(path string) *CookiePathInvalid {
	return &CookiePathInvalid{
		CookieError: &CookieError{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{
				Message: fmt.Sprintf("Can't load the provided cookie file: %s", path),
			},
		},
		Path: path,
	}
}

// CookieInvalid Cookie 无效
type CookieInvalid struct {
	*CookieError
	Path string
}

func NewCookieInvalid(path string) *CookieInvalid {
	return &CookieInvalid{
		CookieError: &CookieError{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{
				Message: fmt.Sprintf("The cookies provided are not valid (may have expired): %s", path),
			},
		},
		Path: path,
	}
}

// CouldNotRetrieveTranscript 无法获取字幕的基类
type CouldNotRetrieveTranscript struct {
	*YouTubeTranscriptApiException
	VideoID string
}

func (e *CouldNotRetrieveTranscript) buildErrorMessage() string {
	videoURL := fmt.Sprintf(WatchURLTemplate, e.VideoID)
	errorMsg := fmt.Sprintf("\nCould not retrieve a transcript for the video %s!", videoURL)

	cause := e.Cause()
	if cause != "" {
		errorMsg += fmt.Sprintf(" This is most likely caused by:\n\n%s", cause)
		errorMsg += "\n\nIf you are sure that the described cause is not responsible for this error " +
			"and that a transcript should be retrievable, please create an issue at " +
			"https://github.com/jdepoix/youtube-transcript-api/issues. " +
			"Please add which version of youtube_transcript_api you are using " +
			"and provide the information needed to replicate the error. " +
			"Also make sure that there are no open issues which already describe your problem!"
	}

	return errorMsg
}

func (e *CouldNotRetrieveTranscript) Cause() string {
	return ""
}

func (e *CouldNotRetrieveTranscript) Error() string {
	return e.buildErrorMessage()
}

// YouTubeDataUnparsable YouTube 数据无法解析
type YouTubeDataUnparsable struct {
	*CouldNotRetrieveTranscript
}

func NewYouTubeDataUnparsable(videoID string) *YouTubeDataUnparsable {
	return &YouTubeDataUnparsable{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *YouTubeDataUnparsable) Cause() string {
	return "The data required to fetch the transcript is not parsable. This should " +
		"not happen, please open an issue (make sure to include the video ID)!"
}

// YouTubeRequestFailed YouTube 请求失败
type YouTubeRequestFailed struct {
	*CouldNotRetrieveTranscript
	Reason string
}

func NewYouTubeRequestFailed(videoID string, err error) *YouTubeRequestFailed {
	return &YouTubeRequestFailed{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
		Reason: err.Error(),
	}
}

func (e *YouTubeRequestFailed) Cause() string {
	return fmt.Sprintf("Request to YouTube failed: %s", e.Reason)
}

// VideoUnplayable 视频无法播放
type VideoUnplayable struct {
	*CouldNotRetrieveTranscript
	Reason     string
	SubReasons []string
}

func NewVideoUnplayable(videoID string, reason string, subReasons []string) *VideoUnplayable {
	return &VideoUnplayable{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
		Reason:     reason,
		SubReasons: subReasons,
	}
}

func (e *VideoUnplayable) Cause() string {
	reason := e.Reason
	if reason == "" {
		reason = "No reason specified!"
	}

	if len(e.SubReasons) > 0 {
		subReasonsText := "\n\nAdditional Details:\n"
		for _, subReason := range e.SubReasons {
			subReasonsText += fmt.Sprintf(" - %s\n", subReason)
		}
		reason += subReasonsText
	}

	return fmt.Sprintf("The video is unplayable for the following reason: %s", reason)
}

// VideoUnavailable 视频不可用
type VideoUnavailable struct {
	*CouldNotRetrieveTranscript
}

func NewVideoUnavailable(videoID string) *VideoUnavailable {
	return &VideoUnavailable{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *VideoUnavailable) Cause() string {
	return "The video is no longer available"
}

// InvalidVideoId 无效的视频 ID
type InvalidVideoId struct {
	*CouldNotRetrieveTranscript
}

func NewInvalidVideoId(videoID string) *InvalidVideoId {
	return &InvalidVideoId{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *InvalidVideoId) Cause() string {
	return "You provided an invalid video id. Make sure you are using the video id and NOT the url!\n\n" +
		"Do NOT run: `YouTubeTranscriptApi().fetch(\"https://www.youtube.com/watch?v=1234\")`\n" +
		"Instead run: `YouTubeTranscriptApi().fetch(\"1234\")`"
}

// RequestBlocked 请求被阻止（IP 封禁）
type RequestBlocked struct {
	*CouldNotRetrieveTranscript
	proxyConfig ProxyConfig
}

func NewRequestBlocked(videoID string) *RequestBlocked {
	return &RequestBlocked{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *RequestBlocked) WithProxyConfig(proxyConfig ProxyConfig) *RequestBlocked {
	e.proxyConfig = proxyConfig
	return e
}

func (e *RequestBlocked) Cause() string {
	baseCause := "YouTube is blocking requests from your IP. This usually is due to one of the " +
		"following reasons:\n" +
		"- You have done too many requests and your IP has been blocked by YouTube\n" +
		"- You are doing requests from an IP belonging to a cloud provider (like AWS, " +
		"Google Cloud Platform, Azure, etc.). Unfortunately, most IPs from cloud " +
		"providers are blocked by YouTube.\n\n"

	if e.proxyConfig != nil {
		if _, ok := e.proxyConfig.(*WebshareProxyConfig); ok {
			return "YouTube is blocking your requests, despite you using Webshare proxies. " +
				"Please make sure that you have purchased \"Residential\" proxies and " +
				"NOT \"Proxy Server\" or \"Static Residential\", as those won't work as " +
				"reliably! The free tier also uses \"Proxy Server\" and will NOT work!\n\n" +
				"The only reliable option is using \"Residential\" proxies (not \"Static " +
				"Residential\"), as this allows you to rotate through a pool of over 30M IPs, " +
				"which means you will always find an IP that hasn't been blocked by YouTube " +
				"yet!\n\n" +
				"You can support the development of this open source project by making your " +
				"Webshare purchases through this affiliate link: " +
				"https://www.webshare.io/?referral_code=w0xno53eb50g \n\n" +
				"Thank you for your support! <3"
		}
		if _, ok := e.proxyConfig.(*GenericProxyConfig); ok {
			return "YouTube is blocking your requests, despite you using proxies. Keep in mind " +
				"that a proxy is just a way to hide your real IP behind the IP of that proxy, " +
				"but there is no guarantee that the IP of that proxy won't be blocked as " +
				"well.\n\n" +
				"The only truly reliable way to prevent IP blocks is rotating through a large " +
				"pool of residential IPs, by using a provider like Webshare " +
				"(https://www.webshare.io/?referral_code=w0xno53eb50g), which provides you " +
				"with a pool of >30M residential IPs (make sure to purchase " +
				"\"Residential\" proxies, NOT \"Proxy Server\" or \"Static Residential\"!).\n\n" +
				"You will find more information on how to easily integrate Webshare here: " +
				"https://github.com/jdepoix/youtube-transcript-api" +
				"?tab=readme-ov-file#using-webshare"
		}
	}

	return baseCause +
		"There are two things you can do to work around this:\n" +
		"1. Use proxies to hide your IP address, as explained in the \"Working around " +
		"IP bans\" section of the README " +
		"(https://github.com/jdepoix/youtube-transcript-api" +
		"?tab=readme-ov-file" +
		"#working-around-ip-bans-requestblocked-or-ipblocked-exception).\n" +
		"2. (NOT RECOMMENDED) If you authenticate your requests using cookies, you " +
		"will be able to continue doing requests for a while. However, YouTube will " +
		"eventually permanently ban the account that you have used to authenticate " +
		"with! So only do this if you don't mind your account being banned!"
}

// IpBlocked IP 被封禁
type IpBlocked struct {
	*RequestBlocked
}

func NewIpBlocked(videoID string) *IpBlocked {
	return &IpBlocked{
		RequestBlocked: NewRequestBlocked(videoID),
	}
}

func (e *IpBlocked) Cause() string {
	return "YouTube is blocking requests from your IP. This usually is due to one of the " +
		"following reasons:\n" +
		"- You have done too many requests and your IP has been blocked by YouTube\n" +
		"- You are doing requests from an IP belonging to a cloud provider (like AWS, " +
		"Google Cloud Platform, Azure, etc.). Unfortunately, most IPs from cloud " +
		"providers are blocked by YouTube.\n\n" +
		"Ways to work around this are explained in the \"Working around IP " +
		"bans\" section of the README (https://github.com/jdepoix/youtube-transcript-api" +
		"?tab=readme-ov-file" +
		"#working-around-ip-bans-requestblocked-or-ipblocked-exception).\n"
}

// TranscriptsDisabled 字幕已禁用
type TranscriptsDisabled struct {
	*CouldNotRetrieveTranscript
}

func NewTranscriptsDisabled(videoID string) *TranscriptsDisabled {
	return &TranscriptsDisabled{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *TranscriptsDisabled) Cause() string {
	return "Subtitles are disabled for this video"
}

// AgeRestricted 年龄限制视频
type AgeRestricted struct {
	*CouldNotRetrieveTranscript
}

func NewAgeRestricted(videoID string) *AgeRestricted {
	return &AgeRestricted{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *AgeRestricted) Cause() string {
	return "This video is age-restricted. Therefore, you are unable to retrieve " +
		"transcripts for it without authenticating yourself.\n\n" +
		"Unfortunately, Cookie Authentication is temporarily unsupported in " +
		"youtube-transcript-api, as recent changes in YouTube's API broke the previous " +
		"implementation. I will do my best to re-implement it as soon as possible."
}

// NotTranslatable 不可翻译
type NotTranslatable struct {
	*CouldNotRetrieveTranscript
}

func NewNotTranslatable(videoID string) *NotTranslatable {
	return &NotTranslatable{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *NotTranslatable) Cause() string {
	return "The requested language is not translatable"
}

// TranslationLanguageNotAvailable 翻译语言不可用
type TranslationLanguageNotAvailable struct {
	*CouldNotRetrieveTranscript
}

func NewTranslationLanguageNotAvailable(videoID string) *TranslationLanguageNotAvailable {
	return &TranslationLanguageNotAvailable{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *TranslationLanguageNotAvailable) Cause() string {
	return "The requested translation language is not available"
}

// FailedToCreateConsentCookie 创建同意 Cookie 失败
type FailedToCreateConsentCookie struct {
	*CouldNotRetrieveTranscript
}

func NewFailedToCreateConsentCookie(videoID string) *FailedToCreateConsentCookie {
	return &FailedToCreateConsentCookie{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *FailedToCreateConsentCookie) Cause() string {
	return "Failed to automatically give consent to saving cookies"
}

// NoTranscriptFound 未找到字幕
type NoTranscriptFound struct {
	*CouldNotRetrieveTranscript
	RequestedLanguageCodes []string
	TranscriptData         *TranscriptList
}

func NewNoTranscriptFound(videoID string, requestedLanguageCodes []string, transcriptData *TranscriptList) *NoTranscriptFound {
	return &NoTranscriptFound{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
		RequestedLanguageCodes: requestedLanguageCodes,
		TranscriptData:         transcriptData,
	}
}

func (e *NoTranscriptFound) Cause() string {
	return fmt.Sprintf("No transcripts were found for any of the requested language codes: %v\n\n%s",
		e.RequestedLanguageCodes, e.TranscriptData.String())
}

// PoTokenRequired 需要 PO Token
type PoTokenRequired struct {
	*CouldNotRetrieveTranscript
}

func NewPoTokenRequired(videoID string) *PoTokenRequired {
	return &PoTokenRequired{
		CouldNotRetrieveTranscript: &CouldNotRetrieveTranscript{
			YouTubeTranscriptApiException: &YouTubeTranscriptApiException{},
			VideoID:                       videoID,
		},
	}
}

func (e *PoTokenRequired) Cause() string {
	return "The requested video cannot be retrieved without a PO Token. If this happens, " +
		"please open a GitHub issue!"
}

// raiseHTTPErrors 检查 HTTP 响应并抛出相应的错误
func raiseHTTPErrors(resp *http.Response, videoID string) error {
	if resp.StatusCode == http.StatusTooManyRequests {
		return NewIpBlocked(videoID)
	}
	if resp.StatusCode >= 400 {
		return NewYouTubeRequestFailed(videoID, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}
	return nil
}
