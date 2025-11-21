package model

import (
	"fmt"

	"github.com/alireza0/x-ui/util/json_util"
	"github.com/alireza0/x-ui/xray"
)

type Protocol string

const (
	VMess       Protocol = "vmess"
	VLESS       Protocol = "vless"
	Dokodemo    Protocol = "Dokodemo-door"
	Http        Protocol = "http"
	Trojan      Protocol = "trojan"
	Shadowsocks Protocol = "shadowsocks"
	Socks       Protocol = "socks"
	Freedom     Protocol = "freedom"
	Blackhole   Protocol = "blackhole"
)

type User struct {
	Id       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Inbound struct {
	Id          int                  `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int                  `json:"-"`
	Up          int64                `json:"up" form:"up"`
	Down        int64                `json:"down" form:"down"`
	Total       int64                `json:"total" form:"total"`
	Remark      string               `json:"remark" form:"remark"`
	Enable      bool                 `json:"enable" form:"enable"`
	ExpiryTime  int64                `json:"expiryTime" form:"expiryTime"`
	ClientStats []xray.ClientTraffic `gorm:"foreignKey:InboundId;references:Id" json:"clientStats" form:"clientStats"`

	// config part
	Listen         string   `json:"listen" form:"listen"`
	Port           int      `json:"port" form:"port"`
	Protocol       Protocol `json:"protocol" form:"protocol"`
	Settings       string   `json:"settings" form:"settings"`
	StreamSettings string   `json:"streamSettings" form:"streamSettings"`
	Tag            string   `json:"tag" form:"tag" gorm:"unique"`
	Sniffing       string   `json:"sniffing" form:"sniffing"`
}

func (i *Inbound) GenXrayInboundConfig() *xray.InboundConfig {
	listen := i.Listen
	if listen != "" {
		listen = fmt.Sprintf("\"%v\"", listen)
	}
	return &xray.InboundConfig{
		Listen:         json_util.RawMessage(listen),
		Port:           i.Port,
		Protocol:       string(i.Protocol),
		Settings:       json_util.RawMessage(i.Settings),
		StreamSettings: json_util.RawMessage(i.StreamSettings),
		Tag:            i.Tag,
		Sniffing:       json_util.RawMessage(i.Sniffing),
	}
}

type Setting struct {
	Id    int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Key   string `json:"key" form:"key"`
	Value string `json:"value" form:"value"`
}

type Client struct {
	ID         string `json:"id"`
	Password   string `json:"password"`
	Flow       string `json:"flow"`
	Email      string `json:"email"`
	TotalGB    int64  `json:"totalGB" form:"totalGB"`
	ExpiryTime int64  `json:"expiryTime" form:"expiryTime"`
	Enable     bool   `json:"enable" form:"enable"`
	TgID       string `json:"tgId" form:"tgId"`
	SubID      string `json:"subId" form:"subId"`
	Reset      int    `json:"reset" form:"reset"`
}

type VLESSSettings struct {
	Clients    []Client `json:"clients"`
	Decryption string   `json:"decryption"`
	Encryption string   `json:"encryption"`
	Fallbacks  []any    `json:"fallbacks"`
}

type Outbound struct {
	Id             int      `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	UserId         int      `json:"-"`
	Protocol       Protocol `json:"protocol" form:"protocol"`
	Address        string   `json:"address" form:"address"`
	Port           int      `json:"port" form:"port"`
	Settings       string   `json:"settings" form:"settings"`
	StreamSettings string   `json:"streamSettings" form:"streamSettings"`
	Tag            string   `json:"tag" form:"tag" gorm:"unique"`
	Remark         string   `json:"remark" form:"remark"`
	Enable         bool     `json:"enable" form:"enable"`
	IsSystem       bool     `json:"isSystem" form:"isSystem" gorm:"default:false"` // true for built-in outbounds like direct/block
	GroupName      string   `json:"groupName" form:"groupName" gorm:"default:'default'"` // group name for organization
	LastTestTime   int64    `json:"lastTestTime" form:"lastTestTime"`
	LastTestResult int      `json:"lastTestResult" form:"lastTestResult"` // latency in ms, -1 for failed
	TestFailCount  int      `json:"testFailCount" form:"testFailCount"`
	CreatedAt      int64    `json:"createdAt" form:"createdAt"`
	UpdatedAt      int64    `json:"updatedAt" form:"updatedAt"`
}

type OutboundTestHistory struct {
	Id           int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	OutboundId   int    `json:"outboundId" form:"outboundId"`
	TestTime     int64  `json:"testTime" form:"testTime"`
	Success      bool   `json:"success" form:"success"`
	LatencyMs    int    `json:"latencyMs" form:"latencyMs"`
	ErrorMessage string `json:"errorMessage" form:"errorMessage"`
}

type Subscription struct {
	Id             int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	GroupName      string `json:"groupName" form:"groupName" gorm:"not null"` // associated group
	Remark         string `json:"remark" form:"remark" gorm:"not null"`       // user-defined name
	URL            string `json:"url" form:"url" gorm:"not null"`             // subscription URL
	Enabled        bool   `json:"enabled" form:"enabled" gorm:"default:true"`
	UpdateInterval int    `json:"updateInterval" form:"updateInterval" gorm:"default:0"` // auto-update interval in hours (0 = manual only)
	UserAgent      string `json:"userAgent" form:"userAgent" gorm:"default:''"`
	LastUpdate     int64  `json:"lastUpdate" form:"lastUpdate"` // unix timestamp of last successful update
	CreatedAt      int64  `json:"createdAt" form:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt" form:"updatedAt"`
}
