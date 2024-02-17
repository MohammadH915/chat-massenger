package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

func handleConnection(conn net.Conn) {
	for {
		dec := gob.NewDecoder(conn)
		m := &Response{}
		err := dec.Decode(m)
		if err != nil {
			return
		}
		println("new Command : ", m.Command)
		if m.Command == "message" {
			println(m.Message.Time, time.Unix(m.Message.Time, 0).String(), "<", m.Message.Channel, ">", m.Message.Sender, ": ", m.Message.Message)
		} else if m.Command == "subScribe" {
			println(m.Channels[0].NLastMessage, "Hour Last Messages : ")
			for _, message := range m.Channels[0].Messages {
				if int(time.Now().Unix()-message.Time) < 3600*m.Channels[0].NLastMessage {
					println(message.Time, time.Unix(message.Time, 0).String(), "<", message.Channel, ">", message.Sender, ": ", message.Message)
				}
			}
		} else if m.Command == "Error" {
			println("Error: ", m.Message.Message)
		} else if m.Command == "History" {
			println("History: ", m.Message.Message)
		} else {
			for _, ch := range m.Channels {
				println("Name:", ch.Name, ",", "Description:", ch.Description)
			}
		}
		println()
	}
}

func main() {
	var username string
	println("Enter username : ")
	fmt.Scan(&username)

	conn, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	go handleConnection(conn)

	for {
		var tp string
		println("enter Type of Command : \n1-Publish\n2-Leave\n3-Subscribe\n4-List\n5-SubList\n6-Create\n7-History")
		fmt.Scan(&tp)
		m := &Msg{
			Username:  username,
			Command:   "",
			Publish:   Publish{},
			Subscribe: Subscribe{},
			Create:    Create{},
			Leave:     Leave{},
		}
		switch tp {
		case "1":
			var channelName string
			println("enter channel Name")
			fmt.Scan(&channelName)
			println("Enter Message")
			var message string
			fmt.Scan(&message)
			m.Publish.Channel = channelName
			m.Publish.Message = message
			m.Command = "publish"
		case "3":
			var channelName string
			println("Enter Channel Name")
			fmt.Scan(&channelName)
			m.Subscribe.Channel = channelName
			m.Command = "subscribe"
		case "2":
			var channelName string
			println("Enter Channel Name")
			fmt.Scan(&channelName)
			m.Leave.Channel = channelName
			m.Command = "leave"
		case "4":
			m.Command = "list"
		case "5":
			m.Command = "sublist"
		case "6":
			var channelName, description string
			println("Enter Channel Name")
			fmt.Scan(&channelName)
			println("Enter Description Name")
			fmt.Scan(&description)
			var nLastMessage int
			println("Enter nLastMessage Name")
			fmt.Scan(&nLastMessage)
			m.Command = "create"
			m.Create.NLastMessage = nLastMessage
			m.Create.Description = description
			m.Create.Channel = channelName
		case "7":
			var channelName, sender string
			println("Enter Channel Name")
			fmt.Scan(&channelName)
			println("Enter sender Name")
			fmt.Scan(&sender)
			var time int64
			println("Enter time")
			fmt.Scan(&time)
			m.Command = "history"
			m.History.Channel = channelName
			m.History.Sender = sender
			m.History.Time = time
		}
		encoder := gob.NewEncoder(conn)
		err := encoder.Encode(m)
		if err != nil {
			fmt.Println("Error broadcasting message:", err)
			return
		}
	}
}
