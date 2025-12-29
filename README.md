# YouTube Transcript API - Go Implementation

[![Go Reference](https://pkg.go.dev/badge/github.com/dogslee/youtube_transcript_api.svg)](https://pkg.go.dev/github.com/dogslee/youtube_transcript_api)
[![Go Report Card](https://goreportcard.com/badge/github.com/dogslee/youtube_transcript_api)](https://goreportcard.com/report/github.com/dogslee/youtube_transcript_api)

This is a Go implementation of the YouTube Transcript API, ported from [jdepoix/youtube-transcript-api](https://github.com/jdepoix/youtube-transcript-api) (Python version).

This project maintains the same API design and feature set as the original Python version, reimplemented in Go for better performance and type safety.

**Documentation:** [pkg.go.dev/github.com/dogslee/youtube_transcript_api](https://pkg.go.dev/github.com/dogslee/youtube_transcript_api)

## Features

- Get available transcript list for YouTube videos
- Download transcript content in specified languages
- Support for both manually created and auto-generated transcripts
- Support for transcript translation
- Multiple output formats (JSON, SRT, WebVTT, plain text)
- Proxy configuration support (generic proxy and Webshare proxy)
- Command-line tool with batch processing support

## Requirements

- **Go version**: 1.19.0 or higher

For detailed compatibility information, please refer to [COMPATIBILITY.md](COMPATIBILITY.md)

## Installation

### As a Library

```bash
go get github.com/dogslee/youtube_transcript_api
```

### Install Command-Line Tool

#### Method 1: Build with Custom Name (Recommended)

Build and install with a custom executable name:

```bash
# Clone the repository
git clone https://github.com/dogslee/youtube_transcript_api.git
cd youtube_transcript_api

# Build and install with custom name
go build -o youtube-transcript-api ./cmd
sudo mv youtube-transcript-api /usr/local/bin/

# Or install to your local bin directory
go build -o ~/go/bin/youtube-transcript-api ./cmd
```

After installation, ensure that `/usr/local/bin` or `~/go/bin` is in your `PATH` environment variable, then you can use it:

```bash
youtube-transcript-api dQw4w9WgXcQ
```

#### Method 2: Using go install (with alias)

Install using `go install` and create an alias:

```bash
go install github.com/dogslee/youtube_transcript_api/cmd@latest

# Create an alias (add to your ~/.bashrc, ~/.zshrc, etc.)
alias youtube-transcript-api='cmd'

# Or create a symbolic link
ln -s $(go env GOPATH)/bin/cmd $(go env GOPATH)/bin/youtube-transcript-api
```

#### Method 3: Using Install Script (Easiest)

Use the provided install script:

```bash
# Clone the repository
git clone https://github.com/dogslee/youtube_transcript_api.git
cd youtube_transcript_api

# Run the install script
./install.sh

# Or install to a custom directory
INSTALL_DIR=~/bin ./install.sh
```

The script will:
- Build the binary with the proper name (`youtube-transcript-api`)
- Install it to the appropriate directory
- Check if the directory is in your PATH
- Provide instructions if PATH needs to be updated

**Uninstallation:**

To uninstall the command-line tool:

```bash
# If installed via Method 1
sudo rm /usr/local/bin/youtube-transcript-api
# Or
rm ~/go/bin/youtube-transcript-api

# If installed via Method 2
rm $(go env GOPATH)/bin/cmd
rm $(go env GOPATH)/bin/youtube-transcript-api  # if symlink was created
```

## Usage Examples

### As a Library

```go
package main

import (
    "fmt"
    yt "github.com/dogslee/youtube_transcript_api"
)

func main() {
    // Create API instance
    api, err := yt.NewYouTubeTranscriptApi(nil)
    if err != nil {
        panic(err)
    }
    
    // Fetch transcript
    transcript, err := api.Fetch("video_id", []string{"en"}, false)
    if err != nil {
        panic(err)
    }
    
    // Print transcript
    for _, snippet := range transcript.Snippets {
        fmt.Printf("[%.2f] %s\n", snippet.Start, snippet.Text)
    }
}
```

### Get Available Transcript List

```go
api, _ := yt.NewYouTubeTranscriptApi(nil)
transcriptList, err := api.List("video_id")
if err != nil {
    panic(err)
}

// Find transcript in specified languages
transcript, err := transcriptList.FindTranscript([]string{"en", "zh"})
if err != nil {
    panic(err)
}

// Fetch transcript content
fetched, err := transcript.Fetch(false)
if err != nil {
    panic(err)
}
```

### Using Proxy

```go
// Generic proxy
proxyConfig, _ := yt.NewGenericProxyConfig("http://proxy.example.com:8080", "")

// Webshare proxy
proxyConfig, _ := yt.NewWebshareProxyConfig(
    "username",
    "password",
    nil, // filterIPLocations
    10,  // retriesWhenBlocked
    "",  // domainName (use default)
    0,   // proxyPort (use default)
)

api, _ := yt.NewYouTubeTranscriptApi(proxyConfig)
```

### Format Output

```go
formatterLoader := yt.NewFormatterLoader()

// JSON format
jsonFormatter, _ := formatterLoader.Load("json")
jsonOutput, _ := jsonFormatter.FormatTranscript(transcript)

// SRT format
srtFormatter, _ := formatterLoader.Load("srt")
srtOutput, _ := srtFormatter.FormatTranscript(transcript)

// WebVTT format
webvttFormatter, _ := formatterLoader.Load("webvtt")
webvttOutput, _ := webvttFormatter.FormatTranscript(transcript)

// Plain text format
textFormatter, _ := formatterLoader.Load("text")
textOutput, _ := textFormatter.FormatTranscript(transcript)
```

## Command-Line Tool

### Installation Methods

**Method 1: Build with Custom Name (Recommended)**

Build and install with a custom executable name:

```bash
# Clone the repository
git clone https://github.com/dogslee/youtube_transcript_api.git
cd youtube_transcript_api

# Build and install
go build -o youtube-transcript-api ./cmd
sudo mv youtube-transcript-api /usr/local/bin/

# Or install to your local bin directory
go build -o ~/go/bin/youtube-transcript-api ./cmd
```

**Method 2: Using go install (with alias)**

Install using `go install` and create an alias or symbolic link:

```bash
go install github.com/dogslee/youtube_transcript_api/cmd@latest

# Create an alias (add to your ~/.bashrc, ~/.zshrc, etc.)
alias youtube-transcript-api='cmd'

# Or create a symbolic link
ln -s $(go env GOPATH)/bin/cmd $(go env GOPATH)/bin/youtube-transcript-api
```

### Usage Examples

```bash
# Fetch transcript
youtube-transcript-api dQw4w9WgXcQ

# List available transcripts
youtube-transcript-api --list-transcripts dQw4w9WgXcQ

# Specify languages
youtube-transcript-api --languages "en zh" dQw4w9WgXcQ

# Specify output format
youtube-transcript-api --format json dQw4w9WgXcQ

# Translate transcript
youtube-transcript-api --translate zh dQw4w9WgXcQ

# Use proxy
youtube-transcript-api --http-proxy "http://proxy.example.com:8080" dQw4w9WgXcQ
```

## API Documentation

### YouTubeTranscriptApi

The main API interface.

#### NewYouTubeTranscriptApi(proxyConfig ProxyConfig) (*YouTubeTranscriptApi, error)

Create a new API instance.

**Parameters:**
- `proxyConfig`: Optional proxy configuration

**Returns:**
- `*YouTubeTranscriptApi`: API instance
- `error`: Error information

#### Fetch(videoID string, languages []string, preserveFormatting bool) (*FetchedTranscript, error)

Fetch transcript for a single video.

**Parameters:**
- `videoID`: Video ID (not the full URL)
- `languages`: List of language codes (ordered by priority)
- `preserveFormatting`: Whether to preserve HTML formatting tags

**Returns:**
- `*FetchedTranscript`: Fetched transcript
- `error`: Error information

#### List(videoID string) (*TranscriptList, error)

Get the list of available transcripts for a video.

**Parameters:**
- `videoID`: Video ID

**Returns:**
- `*TranscriptList`: Transcript list
- `error`: Error information

### TranscriptList

Transcript list object.

#### FindTranscript(languageCodes []string) (*Transcript, error)

Find transcript (preferring manually created ones).

#### FindManuallyCreatedTranscript(languageCodes []string) (*Transcript, error)

Find only manually created transcripts.

#### FindGeneratedTranscript(languageCodes []string) (*Transcript, error)

Find only auto-generated transcripts.

### Transcript

Transcript object.

#### Fetch(preserveFormatting bool) (*FetchedTranscript, error)

Fetch the actual transcript content.

#### Translate(languageCode string) (*Transcript, error)

Translate to the specified language.

### Formatter

Formatter interface.

#### FormatTranscript(transcript *FetchedTranscript) (string, error)

Format a single transcript.

#### FormatTranscripts(transcripts []*FetchedTranscript) (string, error)

Format multiple transcripts.

## Error Handling

All errors implement the `error` interface. Main error types include:

- `CouldNotRetrieveTranscript`: Base class for transcript retrieval failures
- `VideoUnavailable`: Video is unavailable
- `VideoUnplayable`: Video is unplayable
- `TranscriptsDisabled`: Transcripts are disabled
- `NoTranscriptFound`: No transcript found
- `RequestBlocked`: Request blocked (IP banned)
- `AgeRestricted`: Age-restricted video
- And more...

## Notes

1. **Thread Safety**: `YouTubeTranscriptApi` is not thread-safe. In multi-threaded environments, each thread needs to create its own instance.

2. **IP Bans**: YouTube may ban IPs that make frequent requests. It is recommended to use proxies or rotate IPs.

3. **Cookie Authentication**: Cookie authentication is currently not supported, so transcripts for age-restricted videos cannot be retrieved.

4. **API Changes**: YouTube may change its API structure, which may cause some features to fail.

## Acknowledgments

This project is a Go implementation ported from [jdepoix/youtube-transcript-api](https://github.com/jdepoix/youtube-transcript-api).

Thanks to the original project author [@jdepoix](https://github.com/jdepoix) and contributors for creating and maintaining the excellent Python version, which provided important reference and design foundation for this project.

The original project uses the MIT license, and this project also uses the MIT license to maintain license consistency.

## License

This project is licensed under the [MIT License](LICENSE).

Copyright (c) 2025 dogslee

The MIT License is a permissive open-source license that allows you to freely use, modify, and distribute the code, as long as you retain the copyright notice and license text.
