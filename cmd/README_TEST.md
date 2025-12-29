# Testing Guide

## Test Files

- `main_test.go`: Unit tests for configuration and parameter parsing logic
- `integration_test.go`: Integration tests for actual YouTube API calls

## Running Tests

### Run Unit Tests

```bash
cd cmd
go test -v
```

### Run Integration Tests

Integration tests require actual calls to the YouTube API, so a network connection is needed.

```bash
cd cmd
go test -v -tags=integration
```

### Run All Tests (Including Integration Tests)

```bash
cd cmd
go test -v -tags=integration
```

### Skip Integration Tests (Quick Test)

```bash
cd cmd
go test -v -short
```

## Test Content

### Unit Tests (`main_test.go`)

1. **TestCLIConfig**: Tests CLI configuration creation and default values
2. **TestVideoIDSanitization**: Tests video ID sanitization (removing backslashes)
3. **TestExcludeBothFlags**: Tests excluding both manually created and auto-generated transcripts
4. **TestProxyConfig**: Tests proxy configuration
5. **TestLanguageParsing**: Tests language list parsing
6. **TestFlagParsing**: Tests command-line argument parsing

### Integration Tests (`integration_test.go`)

1. **TestIntegration_FetchTranscript**: Tests transcript fetching functionality
2. **TestIntegration_ListTranscripts**: Tests listing available transcripts functionality
3. **TestIntegration_FindTranscript**: Tests finding transcripts in specified languages
4. **TestIntegration_Formatters**: Tests various formatters (JSON, Pretty, Text, SRT, WebVTT)
5. **TestIntegration_CLI**: Tests basic functionality of the command-line tool
6. **TestIntegration_ListTranscriptsCLI**: Tests CLI functionality for listing transcripts
7. **TestIntegration_InvalidVideoID**: Tests error handling for invalid video IDs

## Notes

1. **Network Requirements**: Integration tests require a network connection to access the YouTube API
2. **Test Video**: Integration tests use video ID `jNQXAC9IVRw` (YouTube's first video), which typically has transcripts
3. **Execution Time**: Integration tests may take a long time because they require actual calls to the YouTube API
4. **IP Bans**: Running integration tests frequently may result in IP bans. It is recommended to use proxies or limit test frequency

## Test Coverage

View test coverage:

```bash
cd cmd
go test -tags=integration -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Troubleshooting

If integration tests fail:

1. Check network connection
2. Verify that the test video is still available and has transcripts
3. If encountering IP bans, wait for a while before retrying or use a proxy
4. Check if the YouTube API has changed
