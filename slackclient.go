package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

/* Performs a GET on 'httpGet' and interprets + returns the resulting body as 'response' type */
func HttpGetJsonBody(httpGet string, response interface{}) error {
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

func GetChannels() ([]ChannelInfo, error) {
	var decoded ChannelsResponse
	err := HttpGetJsonBody(HttpGetChannels(), &decoded)
	if err != nil {
		log.Println("Error executing GetChannels", err)
	}
	return decoded.Channels, err
}

func GetChannelHistory(channel string) ([]Message, error) {
	var decoded ChannelHistoryResponse
	err := HttpGetJsonBody(HttpGetChannelHistory(channel), &decoded)
	if err != nil {
		log.Printf("Error executing GetChannelHistory(%v): %v\n", channel, err)
	}
	return decoded.Messages, err
}

func GetUserInfo(user string) (UserInfo, error) {
	var decoded UserInfoResponse
	err := HttpGetJsonBody(HttpGetUserIdentity(user), &decoded)
	if err != nil {
		log.Printf("Error executing HttpGetUserIdentity(%v): %v\n", user, err)
	}
	return decoded.User, err
}
