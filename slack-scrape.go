package main

import (
	"fmt"
	"log"
	"math"
	"sync"
)

// ChannelMember denotes a relation between Slack channel and Slack user.
type ChannelMember struct {
	ChannelID string
	MemberID  string
}

// ChannelMemberInfo represents aggregated information for a ChannelMember.
type ChannelMemberInfo struct {
	Creator      bool
	JoinTime     float64
	MessageCount int
	Left         bool
}

// NewChannelMemberInfo is a factory for new ChannelMemberInfo instances.
func NewChannelMemberInfo() ChannelMemberInfo {
	return ChannelMemberInfo{
		Creator:      false,
		JoinTime:     math.MaxFloat64,
		MessageCount: 0,
		Left:         false,
	}
}

// Summarize allows building a ChannelMemberInfo for a Message.
func (msg Message) Summarize() ChannelMemberInfo {
	return ChannelMemberInfo{
		Creator:      false,
		JoinTime:     math.MaxFloat64,
		MessageCount: 1,
		Left:         false,
	}
}

// Merge allows combining the information within two ChannelMemberInfo instances.
func (cmi ChannelMemberInfo) Merge(other ChannelMemberInfo) ChannelMemberInfo {
	return ChannelMemberInfo{
		Creator:      cmi.Creator || other.Creator,
		JoinTime:     math.Min(cmi.JoinTime, other.JoinTime),
		MessageCount: cmi.MessageCount + other.MessageCount,
		Left:         cmi.Left || other.Left,
	}
}

// AtomicChannelSummaries is supposed to be a concurrency-safe ChannelMember(Info) map.
// It only is if you actually use `mux` though :)
type AtomicChannelSummaries struct {
	v   map[ChannelMember]ChannelMemberInfo
	mux sync.Mutex
}

// UpdateAtomic allows retrieving the current value from an AtomicChannelSummaries and updating it in one go.
func (c *AtomicChannelSummaries) UpdateAtomic(key ChannelMember, f func(ChannelMemberInfo) ChannelMemberInfo) {
	c.mux.Lock()
	c.v[key] = f(c.v[key])
	c.mux.Unlock()
}

// MergeAtomic allows merging a map of ChannelMember(Info)s into an existing AtomicChannelSummaries
func (c *AtomicChannelSummaries) MergeAtomic(m map[ChannelMember]ChannelMemberInfo) {
	for k, v := range m {
		c.UpdateAtomic(k, func(cmi ChannelMemberInfo) ChannelMemberInfo { return cmi.Merge(v) })
	}
}

// SummarizeChannel traverses the history of a channel and updates the given AtomicChannelSummaries accordingly
func SummarizeChannel(channel ChannelInfo, wg *sync.WaitGroup, summary *AtomicChannelSummaries) {
	wg.Add(1)
	go func() {
		// Decrement the counter when the goroutine completes.
		defer wg.Done()

		for elem := range TraverseChannelHistory(channel.Id) {
			if elem.Error != nil {
				log.Fatalf("Failed to retrieve messages for channel %v due to %v\n", channel.Name, elem.Error)
			}
			summary.MergeAtomic(SummarizeMessages(channel.Id, elem.Fragment))
		}
	}()
}

// SummarizeMessages summarizes a bunch Messages from a single channel into a ChannelMember(Info) map
func SummarizeMessages(channelId string, msgs []Message) map[ChannelMember]ChannelMemberInfo {
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
