package him

type Session interface {
	GetChannelId() string
	GetGateId() string
	GetAccount() string
	GetZone() string
	GetIsp() string
	GetRemoteIP() string
	GetDevice() string
	GetApp() string
	GetTags() []string
}
