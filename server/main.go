package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net"
	"strings"
	"time"
)

func findIndex[T comparable](arr []T, target T) int {
	for i, val := range arr {
		if val == target {
			return i
		}
	}
	return -1 // If element is not found, return -1
}

func PubSubChannels(str string) []string {
	str = str[7:(len(str) - 1)]
	if str == "" {
		return []string{}
	}
	str = strings.ReplaceAll(str, " ", "")
	return strings.Split(str, ",")
}

func isInPubSubChannels(cli *Client, check string) bool {
	for _, str := range PubSubChannels(cli.RedisSub.String()) {
		if str == check {
			return true
		}
	}
	return false
}

func getChannelFromRedis(ctx context.Context, redisClient *redis.Client, channelName string) (Channel, error) {
	val, err := redisClient.Get(ctx, channelName).Bytes()
	var ch Channel
	if err != nil {
		return ch, err
	}

	err = json.Unmarshal(val, &ch)
	return ch, nil
}

func isInChannels(ctx context.Context, redisClient *redis.Client, name string) bool {
	channelsName, err := redisClient.PubSubChannels(ctx, "*").Result()
	if err != nil {
		return false
	}
	for _, str := range channelsName {
		if str == name {
			return true
		}
	}
	return false
}

func AddToHistory(ctx context.Context, message ChatMessage, redisHistory *redis.Client) {
	println(fmt.Sprintf("%s-%s-%d", message.Channel, message.Sender, message.Time))
	redisHistory.Set(ctx, fmt.Sprintf("%s-%s-%d", message.Channel, message.Sender, message.Time), message.Message, 0)
}

func handleConnection(ctx context.Context, cli *Client, redisClient *redis.Client, ch chan Response, redisHistory *redis.Client) {
	for {
		dec := gob.NewDecoder(cli.Conn)
		m := &Msg{}
		err := dec.Decode(m)
		if err != nil {
			return
		}
		fmt.Println("Hello: ", m.Username, m.Command)

		switch m.Command {
		case "publish":
			chatMessage := ChatMessage{
				Sender:  m.Username,
				Message: m.Publish.Message,
				Channel: m.Publish.Channel,
				Time:    time.Now().Unix(),
			}
			change, err := getChannelFromRedis(ctx, redisClient, m.Publish.Channel)
			if err != nil || !isInChannels(ctx, redisClient, m.Publish.Channel) {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "Channel Not Found"}}
				continue
			}
			redisHistory.Set(ctx, fmt.Sprintf("%s-%s-%d", chatMessage.Channel, chatMessage.Sender, chatMessage.Time), chatMessage.Message, 0)
			println(redisHistory.Get(ctx, fmt.Sprintf("%s-%s-%d", chatMessage.Channel, chatMessage.Sender, chatMessage.Time)).String())
			redisClient.Publish(ctx, m.Publish.Channel, chatMessage.String())
			change.Messages = append(change.Messages, chatMessage)
			jsonBytes, _ := json.Marshal(change)
			redisClient.Set(ctx, change.Name, jsonBytes, 0)

		case "subscribe":
			err := cli.RedisSub.Subscribe(ctx, m.Subscribe.Channel)
			if err != nil {
				panic("subscribe error")
			}
			change, err := getChannelFromRedis(ctx, redisClient, m.Subscribe.Channel)
			if err != nil || !isInChannels(ctx, redisClient, m.Subscribe.Channel) {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "Channel Not Found"}}
				continue
			}
			if isInPubSubChannels(cli, m.Subscribe.Channel) {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "Already subscribe"}}
				continue
			}
			change.Members = append(change.Members, m.Username)
			jsonBytes, _ := json.Marshal(change)
			redisClient.Set(ctx, change.Name, jsonBytes, 0)
			ch <- Response{Command: "subScribe", Channels: []Channel{change}}

		case "leave":
			err := cli.RedisSub.Unsubscribe(ctx, m.Leave.Channel)
			if err != nil {
				panic("subscribe error")
			}
			change, err := getChannelFromRedis(ctx, redisClient, m.Leave.Channel)
			if err != nil || !isInChannels(ctx, redisClient, m.Leave.Channel) {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "Channel Not Found"}}
				continue
			}
			if !isInPubSubChannels(cli, m.Leave.Channel) {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "Unsubscribe Channel"}}
				continue
			}
			index := findIndex(change.Members, m.Username)
			change.Members = append(change.Members[:index], change.Members[index+1:]...)
			jsonBytes, _ := json.Marshal(change)
			redisClient.Set(ctx, change.Name, jsonBytes, 0)

		case "sublist":
			println(cli.RedisSub.String())
			channels := make([]Channel, len(PubSubChannels(cli.RedisSub.String())))
			for i, str := range PubSubChannels(cli.RedisSub.String()) {
				channels[i], err = getChannelFromRedis(ctx, redisClient, str)
			}
			ch <- Response{Command: "List", Channels: channels}

		case "list":
			channelsName, err := redisClient.PubSubChannels(ctx, "*").Result()
			if err != nil {
				panic(err)
			}
			channels := make([]Channel, len(channelsName))
			for i, str := range channelsName {
				channels[i], err = getChannelFromRedis(ctx, redisClient, str)
			}
			ch <- Response{Command: "List", Channels: channels}

		case "create":
			if isInChannels(ctx, redisClient, m.Create.Channel) == true {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "Channel Is created before"}}
				continue
			}
			newChannel := &Channel{
				Creator:      m.Username,
				Name:         m.Create.Channel,
				CreateTime:   time.Now().Unix(),
				Description:  m.Create.Description,
				Members:      []string{m.Username},
				Messages:     []ChatMessage{},
				NLastMessage: m.Create.NLastMessage,
			}
			err = cli.RedisSub.Subscribe(ctx, m.Create.Channel)
			if err != nil {
				panic("subscribe error")
			}
			jsonBytes, _ := json.Marshal(newChannel)
			redisClient.Set(ctx, newChannel.Name, jsonBytes, 0)

		case "history":
			if !isInChannels(ctx, redisClient, m.History.Channel) {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "channel Not Found"}}
				continue
			}
			val, err := redisHistory.Get(ctx, fmt.Sprintf("%s-%s-%d", m.History.Channel, m.History.Sender, m.History.Time)).Result()
			println(val, fmt.Sprintf("%s-%s-%d", m.History.Channel, m.History.Sender, m.History.Time), err)
			if val == "" || err != nil {
				ch <- Response{Command: "Error", Message: ChatMessage{Message: "Not Message with this information"}}
				continue
			}
			ch <- Response{Command: "History", Message: ChatMessage{Message: val}}

		}
	}
}

func broadcastMessages(ctx context.Context, cli *Client, ch chan Response) {
	go func() {
		for {
			newResponse := <-ch

			encoder := gob.NewEncoder(cli.Conn)
			err := encoder.Encode(newResponse)
			if err != nil {
				fmt.Println("Error broadcasting message:", err)
				return
			}
		}
	}()
	for msg := range cli.RedisSub.Channel() {
		chatMessage := ChatMessage{}
		chatMessage.Fill(msg.Payload)
		println(msg.Payload)
		message := Response{
			Command: "message",
			Message: chatMessage,
		}
		encoder := gob.NewEncoder(cli.Conn)
		err := encoder.Encode(message)
		if err != nil {
			fmt.Println("Error broadcasting message:", err)
			return
		}
	}
}

func main() {
	ctx := context.Background()
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer redisClient.Close()

	redisHistory := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})
	defer redisHistory.Close()

	fmt.Println("Server is running on port 1234...")

	connections := make(map[*Client]bool)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		cli := &Client{
			Conn:     conn,
			RedisSub: redisClient.Subscribe(ctx),
		}
		connections[cli] = true
		ch := make(chan Response)
		go broadcastMessages(ctx, cli, ch)
		go handleConnection(ctx, cli, redisClient, ch, redisHistory)
	}
}
