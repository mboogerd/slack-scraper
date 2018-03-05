package main

import "fmt"

const (
	// Channels API endpoint lists all the channels in your Slack workspace
	Channels = "conversations.list"
	// ChannelHistory API endpoint retrieves the full message history for a given channel
	ChannelHistory = "conversations.history"
	// UsersInfo API endpoint retrieves user details for a given user
	UsersInfo = "users.info"
)

// LimitOfCursorTraversal configures how many elements are retrieved in once when using a cursor-traversal of a Slack API
const LimitOfCursorTraversal = 1

// ChannelInfo contains the principal information in a Slack channel
type ChannelInfo struct {
	Id          string
	Name        string
	Creator     string
	Is_channel  bool
	Is_group    bool
	Is_im       bool
	Is_private  bool
	Created     int
	Is_archived bool
	Is_general  bool
}

// ChannelsResponse represents the response from the conversations.list API
type ChannelsResponse struct {
	Ok                bool
	Channels          []ChannelInfo
	Response_metadata ResponseMetadata
}

// Message contains the principal information of a Slack message
type Message struct {
	Type    string
	Subtype string
	User    string
	Text    string
	Ts      string
}

// ChannelHistoryResponse represents the response from the conversations.history API
type ChannelHistoryResponse struct {
	Ok                bool
	Messages          []Message
	Response_metadata ResponseMetadata
}

// ProfileData contains the information from the `profile` attribute of the users.info API
type ProfileData struct {
	Real_name            string
	Display_name         string
	Real_name_normalized string
	Image_24             string
	Image_32             string
	Image_48             string
	Image_72             string
	Image_192            string
	Image_512            string
}

// UserInfo contains the principal user information
type UserInfo struct {
	Id       string
	Name     string
	Profile  ProfileData
	Is_admin bool
	Is_owner bool
	Is_bot   bool
	Deleted  bool
}

// UserInfoResponse represents the response from the users.info API
type UserInfoResponse struct {
	Ok   bool
	User UserInfo
}

// ResponseMetadata is used for APIs that support cursor-ed traversal
type ResponseMetadata struct {
	Next_cursor string
}

// ChannelsAPI returns the URL to retrieve all the channels in your workspace
func ChannelsAPI() string {
	return SlackAPI + Channels + "?token=" + token
}

// CursoredChannelsAPI returns the URL to do a cursored traversal of the channels in your workspace
func CursoredChannelsAPI(cursor string) string {
	var baseURL = ChannelsAPI()
	if cursor == "" {
		return baseURL
	}
	return baseURL + "&cursor=" + cursor
}

// ChannelHistoryAPI returns the URL to retrieve the messages of a given channel
func ChannelHistoryAPI(channel string) string {
	return SlackAPI + ChannelHistory + "?token=" + token + "&channel=" + channel + "&limit=" + fmt.Sprint(LimitOfCursorTraversal)
}

// HttpGetChannelHistoryCursor returns the URL to do a cursored traversal of the messages of a given channel
func CursoredChannelHistoryAPI(channel string, cursor string) string {
	var baseURL = ChannelHistoryAPI(channel)
	if cursor == "" {
		return baseURL
	}
	return baseURL + "&cursor=" + cursor
}

// UserIdentityAPI returns the URL to use for the users.info API
func UserIdentityAPI(user string) string {
	return SlackAPI + UsersInfo + "?token=" + token + "&user=" + user
}
