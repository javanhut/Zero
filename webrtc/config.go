package webrtc

import (
	"github.com/pion/webrtc/v4"
)

type Config struct {
	ICEServers []webrtc.ICEServer
}

func DefaultConfig() *Config {
	return &Config{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{
				URLs: []string{"stun:stun1.l.google.com:19302"},
			},
		},
	}
}

func (c *Config) ToWebRTCConfig() webrtc.Configuration {
	return webrtc.Configuration{
		ICEServers: c.ICEServers,
	}
}

func NewConfig(iceServers []webrtc.ICEServer) *Config {
	return &Config{
		ICEServers: iceServers,
	}
}
