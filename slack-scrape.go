package main

import (
	"fmt"
	"log"
	"math"
	"sync"
)

type ChannelMember struct {
	ChannelId string
	MemberId  string
}

type ChannelMemberInfo struct {
	Creator      bool
	JoinTime     float64
	MessageCount int
	Left         bool
}

func NewChannelMemberInfo() ChannelMemberInfo {
	return ChannelMemberInfo{
		Creator:      false,
		JoinTime:     math.MaxFloat64,
		MessageCount: 0,
		Left:         false,
	}
}

func (msg Message) Summarize() ChannelMemberInfo {
	return ChannelMemberInfo{
		Creator:      false,
		JoinTime:     math.MaxFloat64,
		MessageCount: 1,
		Left:         false,
	}
}

func (cmi ChannelMemberInfo) Merge(other ChannelMemberInfo) ChannelMemberInfo {
	return ChannelMemberInfo{
		Creator:      cmi.Creator || other.Creator,
		JoinTime:     math.Min(cmi.JoinTime, other.JoinTime),
		MessageCount: cmi.MessageCount + other.MessageCount,
		Left:         cmi.Left || other.Left,
	}
}

type AtomicChannelSummaries struct {
	v   map[ChannelMember]ChannelMemberInfo
	mux sync.Mutex
}

func (c *AtomicChannelSummaries) UpdateAtomic(key ChannelMember, f func(ChannelMemberInfo) ChannelMemberInfo) {
	c.mux.Lock()
	c.v[key] = f(c.v[key])
	c.mux.Unlock()
}

func (c *AtomicChannelSummaries) MergeAtomic(m map[ChannelMember]ChannelMemberInfo) {
	for k, v := range m {
		c.UpdateAtomic(k, func(cmi ChannelMemberInfo) ChannelMemberInfo { return cmi.Merge(v) })
	}
}

var channelSummaries = AtomicChannelSummaries{v: make(map[ChannelMember]ChannelMemberInfo)}

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
			summary.MergeAtomic(SummarizePartialMessages(channel.Id, elem.Fragment))
		}
	}()
}

func SummarizePartialMessages(channelId string, msgs []Message) map[ChannelMember]ChannelMemberInfo {
	result := make(map[ChannelMember]ChannelMemberInfo)
	for _, msg := range msgs {
		cm := ChannelMember{channelId, msg.User}
		cmi, err := result[cm]
		if err == false {
			cmi = NewChannelMemberInfo()
		}
		result[cm] = cmi.Merge(msg.Summarize())
	}
	return result
}
