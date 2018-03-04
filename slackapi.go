package main

const (
	Channels       = "conversations.list"
	ChannelHistory = "conversations.history"
	UsersInfo      = "users.info"
)

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

type ChannelsResponse struct {
	Ok       bool
	Channels []ChannelInfo
}

type Message struct {
	Type    string
	Subtype string
	User    string
	Text    string
	Ts      string
}

type ChannelHistoryResponse struct {
	Ok       bool
	Messages []Message
}

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

type UserInfo struct {
	Id       string
	Name     string
	Profile  ProfileData
	Is_admin bool
	Is_owner bool
	Is_bot   bool
	Deleted  bool
}

type UserInfoResponse struct {
	Ok   bool
	User UserInfo
}

func HttpGetChannels() string {
	return SlackApi + Channels + "?token=" + token
}

func HttpGetChannelHistory(channel string) string {
	return SlackApi + ChannelHistory + "?token=" + token + "&channel=" + channel
}

func HttpGetUserIdentity(user string) string {
	return SlackApi + UsersInfo + "?token=" + token + "&user=" + user
}
