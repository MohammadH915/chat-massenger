package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	Conn     net.Conn
	RedisSub *redis.PubSub
}

type Publish struct {
	Channel string
	Message string
}

type Subscribe struct {
	Channel string
}

type Create struct {
	Channel      string
	Description  string
	NLastMessage int
}

type Leave struct {
	Channel string
}

type History struct {
	Channel string
	Sender  string
	Time    int64
}

type Msg struct {
	Username  string
	Command   string
	Publish   Publish
	Subscribe Subscribe
	Create    Create
	Leave     Leave
	History   History
}

type ChatMessage struct {
	Sender  string
	Message string
	Channel string
	Time    int64
}

func (m *ChatMessage) String() string {
	return fmt.Sprintf("%s$%s$%s$%s", m.Sender, m.Message, m.Channel, strconv.FormatInt(m.Time, 10))
}

func (m *ChatMessage) Fill(str string) {
	spl := strings.Split(str, "$")
	m.Sender = spl[0]
	m.Message = spl[1]
	m.Channel = spl[2]
	m.Time, _ = strconv.ParseInt(spl[3], 10, 64)

}

type Channel struct {
	Creator      string
	Name         string
	CreateTime   int64
	Description  string
	Members      []string
	Messages     []ChatMessage
	NLastMessage int
}

type Response struct {
	Command  string
	Message  ChatMessage
	Channels []Channel
}
