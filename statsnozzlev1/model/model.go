package model

import (
	"fmt"
	"time"
)

type AppSpaceOrg struct {
	AppName   string
	AppGuid   string
	SpaceName string
	SpaceGuid string
	OrgName   string
	OrgGuid   string
}

type Stats struct {
	Time          time.Time
	IP            string
	PeerType      string
	Method        string
	StatusCode    int
	ContentLength int
	URI           string
	Remote        string
	RemotePort    string
	ForwardedFor  string
	UserAgent     string
}

func (stats Stats) String() string {
	return fmt.Sprintf("Time: %s, RTR IP:%s, PeerType:%s, Remote:%s:%s, Method:%s, Uri:%s, StatusCode:%d, ContentLength:%d, Forwarded: %v, UserAgent:%s",
		stats.Time.Format(time.RFC3339), stats.IP, stats.PeerType, stats.Remote, stats.RemotePort, stats.Method, stats.URI, stats.StatusCode, stats.ContentLength, stats.ForwardedFor, stats.UserAgent)
}
