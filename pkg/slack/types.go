package slack

import "strconv"

type Timestamp string

func (t Timestamp) Float64() float64 {
	v, _ := strconv.ParseFloat(string(t), 64)
	return v
}

type ChannelType string

const (
	Public  ChannelType = "public_channel"
	Private ChannelType = "private_channel"
	DM      ChannelType = "im"
	GroupDM ChannelType = "mpim"
)
