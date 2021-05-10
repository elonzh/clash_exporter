package main

import (
	"net"
	"time"
)

type Version struct {
	Premium bool   `json:"premium"`
	Version string `json:"version"`
}

// see clash/constant/adapters.go#AdapterType.String
var (
	ConnectionProxyTypes   = []string{"Shadowsocks", "ShadowsocksR", "Snell", "Socks5", "Http", "Vmess", "Trojan"}
	RuleProxyTypes         = []string{"Direct", "Reject", "Relay", "Selector", "Fallback", "URLTest", "LoadBalance"}
	AllProxyTypes          []string
	connectionProxyTypeSet = make(map[string]struct{}, len(ConnectionProxyTypes))
)

func init() {
	AllProxyTypes = append(AllProxyTypes, ConnectionProxyTypes...)
	AllProxyTypes = append(AllProxyTypes, RuleProxyTypes...)

	for _, t := range ConnectionProxyTypes {
		connectionProxyTypeSet[t] = struct{}{}
	}
}

func IsConnectionProxy(proxy *Proxy) bool {
	_, ok := connectionProxyTypeSet[proxy.Type]
	return ok
}

type Proxy struct {
	Type string
	Name string
	Now  string
	All  []string
}

type Metadata struct {
	NetWork  string `json:"network"`
	Type     string `json:"type"`
	SrcIP    net.IP `json:"sourceIP"`
	DstIP    net.IP `json:"destinationIP"`
	SrcPort  string `json:"sourcePort"`
	DstPort  string `json:"destinationPort"`
	AddrType int    `json:"-"`
	Host     string `json:"host"`
}

type TrackerInfo struct {
	UUID          string    `json:"id"`
	Metadata      *Metadata `json:"metadata"`
	UploadTotal   int64     `json:"upload"`
	DownloadTotal int64     `json:"download"`
	Start         time.Time `json:"start"`
	Chain         []string  `json:"chains"`
	Rule          string    `json:"rule"`
	RulePayload   string    `json:"rulePayload"`
}

type Snapshot struct {
	DownloadTotal int64          `json:"downloadTotal"`
	UploadTotal   int64          `json:"uploadTotal"`
	Connections   []*TrackerInfo `json:"connections"`
}
