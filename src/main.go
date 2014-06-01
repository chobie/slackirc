package main

import (
	"configuration"
	"flag"
	"fmt"
	"net/url"
	"bytes"
	"strconv"
	"net/http"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
)

const (
	libraryVersion = "0.1"
	userAgent = "go-slackic/" + libraryVersion
)

var config *configuration.Config

func SlackWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var s *configuration.Server
	var c *configuration.Channel

	r.ParseForm()

	user_name := r.FormValue("user_name")
	room := r.FormValue("channel_name")
	text := r.FormValue("text")
	channel_token := r.FormValue("token")

	if user_name == "" {
		fmt.Printf("user name not found")
		return
	}
	if room == "" {
		fmt.Printf("room name not found")
		return
	}
	if text == "" {
		fmt.Printf("textn ot found")
		return
	}

	ok := false
	for _, nick := range config.Slack.AcceptNicks {
		if nick == user_name {
			ok = true
		}
	}
	if !ok {
		fmt.Printf("[web server] not accept nick %s\n", user_name)
		return
	}

	for j, server := range config.Servers {
		for i, ch := range server.Channels {
			if ch.SlackChannel == "#" + room {
				if ch.ReadOnly {
					return
				}

				s = &config.Servers[j]
				c = &config.Servers[j].Channels[i]
				break
			}
		}
	}

	if c != nil {
		if c.SlackChannelToken != "" && c.SlackChannelToken != channel_token {
			fmt.Printf("channel token mismatched")
			return
		}

		fmt.Printf("Channel: %v\n", c)
		fmt.Printf("Server: %+v\n", s)
		s.IRC.Privmsg(c.Channel, text)
	} else {
		fmt.Printf("Channel mapping not found for %s\n", c)
		return
	}

	fmt.Printf("Values: %s, %s", user_name, room)
}

func SendMessage(room, message string) {
	data := url.Values{}
	data.Add("token", config.Slack.Token)
	data.Add("channel", room)
	baseUrl, _ := url.Parse(config.Slack.Endpoint)
	url := fmt.Sprintf("%v", baseUrl)

	req, _ := http.NewRequest("POST", url+"?"+data.Encode(), bytes.NewBufferString(message))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(message)))
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)
	fmt.Println(resp)
}


func forever(config *configuration.Config) {
	for offset, server := range config.Servers {
			conf := &config.Servers[offset]
			conn := irc.IRC(server.Nick, server.Name)
			if server.Pass != "" {
				conn.Password = server.Pass
			}

			err := conn.Connect(fmt.Sprintf("%s:%s", server.Host, server.Port))
			if err != nil {
				fmt.Printf("Failed to connect: %s\n", server.Host)
				return
			}
			config.Servers[offset].IRC = conn
			if server.NickServePass != "" {
				conn.SendRaw(fmt.Sprintf("/msg NickServe IDENTIFY %s %s", server.Nick, server.NickServePass))
			}

			for _, channel := range server.Channels {
				conn.Join(channel.Channel)
			}

			conn.AddCallback("PRIVMSG", func(e *irc.Event) {
					// confからChannelを引っ張る
					fmt.Printf("Message [%+v]\n", e.Arguments)
					ok := false

					for _, c := range conf.Channels {
						if c.Channel == e.Arguments[0] {
							message := fmt.Sprintf("<%s> %s", e.Nick, e.Message())
							SendMessage(c.SlackChannel, message)
							ok = true
							break;
						}
					}
					if !ok {
						fmt.Printf("Room not found: %+v", conf.Channels)
					}
				})
			conn.AddCallback("NOTICE", func(e *irc.Event) {
					fmt.Printf("Message [%+v]\n", e.Arguments)
					ok := false

					if e.Arguments[0][0:1] != "#" {
						return
					}

					for _, c := range conf.Channels {
						if c.Channel == e.Arguments[0] {
							message := fmt.Sprintf("<%s> %s", e.Nick, e.Message())
							SendMessage(c.SlackChannel, message)
							break;
						}
					}
					if !ok {
						fmt.Printf("Room not found: %+v", conf.Channels)
					}
				})
			go conn.Loop()
	}
}


func main() {
	configFile := flag.String("config", "config.toml", "the config file")
	flag.Parse()

	config, _ = configuration.LoadConfiguration(*configFile)

	http.HandleFunc("/slack/api/webhooks", SlackWebhookHandler)
	go http.ListenAndServe(fmt.Sprintf("%s:%s", config.Http.Host, config.Http.Port), nil)
	go forever(config)
	select{}
}
