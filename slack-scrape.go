package main

import (
	"fmt"
	"log"
)

func main() {
	channels, err := GetChannels()

	if err != nil {
		log.Fatalln("Error retrieving channels", err)
	}

	for _, channel := range channels {
		user, err := GetUserInfo(channel.Creator)
		if err != nil {
			log.Printf("Failed to retrieve creator %v for channel %v due to %v\n", channel.Creator, channel.Name, err)
		}
		fmt.Printf("CHANNEL [%v]: %v. Created by: %v\n", channel.Id, channel.Name, user.Profile.Real_name_normalized)

		messages, err := GetChannelHistory(channel.Id)
		if err != nil {
			log.Printf("Failed to retrieve messages for channel %v due to %v\n", channel.Name, err)
		}

		for _, message := range messages {

			user, err := GetUserInfo(message.User)
			if err != nil {
				log.Printf("Failed to retrieve username for message of user-id %v due to %v\n", message.User, err)
			}

			fmt.Printf("  [%v] %v %v: %v\n", message.Ts, message.Subtype, user.Profile.Real_name_normalized, message.Text)
		}
	}
}
