package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

const rateLimit = 100 * time.Millisecond
const burstLimit = 3

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
func SummarizeChannel(session SlackSession, channel ChannelInfo, wg *sync.WaitGroup, summary *AtomicChannelSummaries, rateChannel <-chan time.Time) {
	wg.Add(1)
	go func() {
		// Decrement the counter when the goroutine completes.
		defer wg.Done()

		for elem := range TraverseChannelHistory(session, channel.Id) {
			<-rateChannel
			if elem.Error != nil {
				log.Fatalf("Failed to retrieve messages for channel %v due to %v\n", channel.Name, elem.Error)
			}
			summary.MergeAtomic(SummarizeMessages(channel.Id, elem.Fragment))
		}
	}()
}

// SummarizeMessages summarizes a bunch Messages from a single channel into a ChannelMember(Info) map
func SummarizeMessages(channelID string, msgs []Message) map[ChannelMember]ChannelMemberInfo {
	result := make(map[ChannelMember]ChannelMemberInfo)
	for _, msg := range msgs {
		cm := ChannelMember{channelID, msg.User}
		cmi, err := result[cm]
		if err == false {
			cmi = NewChannelMemberInfo()
		}
		result[cm] = cmi.Merge(msg.Summarize())
	}
	return result
}

func main() {
	session := SlackSession{
		API:   os.Getenv("SlackAPI"),
		Token: os.Getenv("SlackToken"),
	}

	var channelSummaries = AtomicChannelSummaries{v: make(map[ChannelMember]ChannelMemberInfo)}

	go startScraper(session, &channelSummaries)
	setupHealthChecks(&channelSummaries)
}

func startScraper(session SlackSession, channelSummaries *AtomicChannelSummaries) {
	rateChannel := setupBurstRateLimiter(rateLimit, burstLimit)

	var channelHistoriesWaitGroup sync.WaitGroup
	for channelsFragment := range TraverseChannels(session) {
		if channelsFragment.Error != nil {
			log.Fatalln("Error retrieving channels", channelsFragment.Error)
		}

		for _, channel := range channelsFragment.Fragment {
			// Don't overwhelm the server
			<-rateChannel
			user, err := GetUserInfo(session, channel.Creator)
			if err != nil {
				log.Printf("Failed to retrieve creator %v for channel %v due to %v\n", channel.Creator, channel.Name, err)
			}
			fmt.Printf("CHANNEL [%v]: %v. Created by: %v\n", channel.Id, channel.Name, user.Profile.Real_name_normalized)

			SummarizeChannel(session, channel, &channelHistoriesWaitGroup, channelSummaries, rateChannel)
		}
	}
	channelHistoriesWaitGroup.Wait()
}

func setupBurstRateLimiter(interval time.Duration, burstLimit int) <-chan time.Time {
	rateChannel := make(chan time.Time, burstLimit)

	go func() {
		// Dispatch an initial 'burstLimit' elements to the limiter channel
		for i := 0; i < burstLimit; i++ {
			rateChannel <- time.Now()
		}
		// Dispatch consecutive signals at a fixed 'rateLimit' rate
		for t := range time.Tick(interval) {
			rateChannel <- t
		}
	}()

	return rateChannel
}
