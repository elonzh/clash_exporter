package main

import (
	"math"
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

const MaxDelay = math.MaxUint16

type ProxyDelay struct {
	Time  time.Time `json:"time"`
	Delay uint16    `json:"delay"`
}

type Proxy struct {
	Type    string
	Name    string
	Now     string
	All     []string
	History []*ProxyDelay `json:"history"`
}

const (
	VehicleTypeFile       = "File"
	VehicleTypeHTTP       = "HTTP"
	VehicleTypeCompatible = "Compatible"
)

type Provider struct {
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	VehicleType string    `json:"vehicleType"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Proxies     []*Proxy  `json:"proxies"`
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
