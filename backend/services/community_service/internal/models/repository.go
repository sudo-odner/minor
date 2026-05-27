package models

type ChannelType int

const (
	ChannelTypeCategory ChannelType = 0
	ChannelTypeText     ChannelType = 1
	ChannelTypeVoice    ChannelType = 2
)

type OverrideType string

const (
	OverrideTypeRole OverrideType = "role"
	OverrideTypeUser OverrideType = "user"
)
