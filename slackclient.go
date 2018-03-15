package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// PartialChannels is either some ChannelInfo elements or an error. TEST
type PartialChannels struct {
	Fragment []ChannelInfo
	Error    error
}

// PartialMessages is either some Message elements or an error
type PartialMessages struct {
	Fragment []Message
	Error    error
}

// HTTPGetJSONBody performs a GET on 'httpGet' and interprets + returns the resulting body as 'response', or returns an error
func HTTPGetJSONBody(httpGet string, response interface{}) error {
	// fetch url
	resp, err := http.Get(httpGet)
	if err != nil {
		log.Printf("[%v] Error fetching: %v\n", httpGet, err)
	}
	// defer response close
	defer resp.Body.Close()

	// confirm we received an OK status
	if resp.StatusCode != http.StatusOK {
		log.Printf("[%v] Error Status not OK: %v\n", httpGet, resp.StatusCode)
	}

	// read the entire body of the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%v] Error reading body: %v\n", httpGet, err)
	}

	// create an empty instance of struct
	// this is what gets filled in when unmarshaling JSON
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("[%v] Error decoing JSON: %v\n", httpGet, err)
	}

	return err
}

// GetChannels invokes the Slack conversation.list API and returns the result or an error
func GetChannels(session SlackSession) ([]ChannelInfo, error) {
	var decoded ChannelsResponse
	err := HTTPGetJSONBody(ChannelsAPI(session), &decoded)
	if err != nil {
		log.Println("Error executing GetChannels", err)
	}
	return decoded.Channels, err
}

// TraverseChannels traverses the Slack conversation.list API through a cursor and returns the result as channel
func TraverseChannels(session SlackSession) <-chan PartialChannels {
	rc := make(chan PartialChannels)
	go func() {
		var decoded ChannelsResponse
		err := HTTPGetJSONBody(CursoredChannelsAPI(session, ""), &decoded)
		if err != nil {
			log.Println("Error executing GetChannels", err)
		}
		for decoded.Response_metadata.Next_cursor != "" {
			rc <- PartialChannels{Fragment: decoded.Channels, Error: err}
			err = HTTPGetJSONBody(CursoredChannelsAPI(session, decoded.Response_metadata.Next_cursor), &decoded)
		}
		rc <- PartialChannels{Fragment: decoded.Channels, Error: err}
		close(rc)
	}()
	return rc
}

// GetChannelHistory invokes the Slack conversation.history API and returns the result or an error
func GetChannelHistory(session SlackSession, channel string, cursor string) (ChannelHistoryResponse, error) {
	var decoded ChannelHistoryResponse
	err := HTTPGetJSONBody(CursoredChannelHistoryAPI(session, channel, cursor), &decoded)
	if err != nil {
		log.Printf("Error executing GetChannelHistory(%v): %v\n", channel, err)
	}
	return decoded, err
}

// TraverseChannelHistory traverses the Slack conversation.history API through a cursor and returns the result as channel
func TraverseChannelHistory(session SlackSession, channel string) <-chan PartialMessages {
	rc := make(chan PartialMessages)
	go func() {
		response, err := GetChannelHistory(session, channel, "")
		for response.Response_metadata.Next_cursor != "" {
			rc <- PartialMessages{Fragment: response.Messages, Error: err}
			response, err = GetChannelHistory(session, channel, response.Response_metadata.Next_cursor)
		}
		rc <- PartialMessages{Fragment: response.Messages, Error: err}
		close(rc)
	}()
	return rc
}

// GetUserInfo obtains the UserInfo data for a given `user`
func GetUserInfo(session SlackSession, user string) (UserInfo, error) {
	var decoded UserInfoResponse
	err := HTTPGetJSONBody(UserIdentityAPI(session, user), &decoded)
	if err != nil {
		log.Printf("Error executing HttpGetUserIdentity(%v): %v\n", user, err)
	}
	return decoded.User, err
}
