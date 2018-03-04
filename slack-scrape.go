package main

import (
	"fmt"
	"log"
	"sync"
)

// TODO: we could explore
type AtomicChannelSummaries struct {
	v   map[string][]Message
	mux sync.Mutex
}

func (c *AtomicChannelSummaries) UpdateAtomic(key string, f func([]Message) []Message) {
	c.mux.Lock()
	c.v[key] = f(c.v[key])
	c.mux.Unlock()
}

var channelSummaries = AtomicChannelSummaries{v: make(map[string][]Message)}

func main() {
	channels, err := GetChannels()

	if err != nil {
		log.Fatalln("Error retrieving channels", err)
	}

	var channelHistoriesWaitGroup sync.WaitGroup
	for _, channel := range channels {

		user, err := GetUserInfo(channel.Creator)
		if err != nil {
			log.Printf("Failed to retrieve creator %v for channel %v due to %v\n", channel.Creator, channel.Name, err)
		}
		fmt.Printf("CHANNEL [%v]: %v. Created by: %v\n", channel.Id, channel.Name, user.Profile.Real_name_normalized)

		SummarizeChannel(channel, &channelHistoriesWaitGroup, &channelSummaries)

		// messages, err := GetChannelHistory(channel.Id)
		// if err != nil {
		// 	log.Printf("Failed to retrieve messages for channel %v due to %v\n", channel.Name, err)
		// }

		// for _, message := range messages {

		// 	user, err := GetUserInfo(message.User)
		// 	if err != nil {
		// 		log.Printf("Failed to retrieve username for message of user-id %v due to %v\n", message.User, err)
		// 	}

		// 	fmt.Printf("  [%v] %v %v: %v\n", message.Ts, message.Subtype, user.Profile.Real_name_normalized, message.Text)
		// }
	}
	channelHistoriesWaitGroup.Wait()

	fmt.Printf("HISTORIES %v\n", &channelSummaries)
}

func SummarizeChannel(channel ChannelInfo, wg *sync.WaitGroup, summary *AtomicChannelSummaries) {
	wg.Add(1)
	go func() {
		// Decrement the counter when the goroutine completes.
		defer wg.Done()

		for elem := range TraverseChannelHistory(channel.Id) {
			if elem.Error != nil {
				log.Fatalf("Failed to retrieve messages for channel %v due to %v\n", channel.Name, elem.Error)
			}
			summary.UpdateAtomic(channel.Id, func(prev []Message) []Message { return append(prev, elem.Fragment...) })
		}
	}()
}
