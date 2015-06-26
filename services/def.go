package services

type ServiceType string

const (
	SERVICE_SNOWFLAKE  = ServiceType("/backends/snowflake")
	SERVICE_GEOIP      = ServiceType("/backends/geoip")
	SERVICE_WORDFILTER = ServiceType("/backends/wordfilter")
	SERVICE_BGSAVE     = ServiceType("/backends/bgsave")
	SERVICE_AUTH       = ServiceType("/backends/auth")
	SERVICE_CHAT       = ServiceType("/backends/chat")
	SERVICE_GAME       = ServiceType("/backends/game_server")
)
