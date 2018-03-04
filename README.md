# README

This is a basic Slack scraper, built as a so-called "integrated API". This essentially means that the client is using a token-based instead of OAuth implementation
for authentication.

## Functionality

* Retrieve all channels from a Slack space
* Retrieve full message history from each channel
* History retrieval is parallellized per channel
* History retrieval is implemented as a sequence of calls over a cursor

### TODO

* Implement rate limiting
* Aggregate interesting elements per channel and participant

## Configuration

Create a go file for configuration, say 'config.go' and set the following two constants

```go
const (
    // Obtain a token from https://api.slack.com/custom-integrations/legacy-tokens
    token = "xoxp-some-user-token-here"
    SlackApi = "https://yourspace.slack.com/api/"
)
```

## Running

go run *.go