package youtube_transcript_api

// YouTube API 相关常量
const (
	WatchURLTemplate        = "https://www.youtube.com/watch?v=%s"
	InnertubeAPIURLTemplate = "https://www.youtube.com/youtubei/v1/player?key=%s"
)

// InnertubeContext 是调用 YouTube InnerTube API 时使用的客户端上下文
var InnertubeContext = map[string]interface{}{
	"context": map[string]interface{}{
		"client": map[string]interface{}{
			"clientName":    "ANDROID",
			"clientVersion": "20.10.38",
		},
	},
}
