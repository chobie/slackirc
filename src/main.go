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
)

const (
	libraryVersion = "0.1"
	userAgent = "go-slack/" + libraryVersion
)

var config *configuration.Config

func SlackWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var s *configuration.Server
	var c *configuration.Channel

	r.ParseForm()

	user_name := r.FormValue("user_name")
	room := r.FormValue("channel_name")
	text := r.FormValue("text")

	if user_name == "" {
		return
	}
	if room == "" {
		return
	}
	if text == "" {
		return
	}

	ok := false
	for _, nick := range config.Slack.AcceptNicks {
		if nick == user_name {
			ok = true
		}
	}
	if !ok {
		fmt.Printf("not accept nick")
		return
	}

	for _, server := range config.Servers {
		for _, ch := range server.Channels {
			if ch.SlackChannel == "#" + room {
				if ch.ReadOnly {
					return
				}

				s = &server
				c = &ch
				break
			}
		}
	}

	if c != nil {
		fmt.Printf("Channel: %v\n", c)
		go s.IRC.Privmsg(c.Channel, text)
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
	fmt.Println(resp)
}


func forever(config *configuration.Config) {
	for offset, server := range config.Servers {
		func() {
			conf := &server
			conn := irc.IRC(server.Nick, server.Name)
			err := conn.Connect(fmt.Sprintf("%s:%s", server.Host, server.Port))
			if err != nil {
				fmt.Printf("Failed to connect: %s\n", server.Host)
				return
			}

			for _, channel := range server.Channels {
				conn.Join(channel.Channel)
			}

			conn.AddCallback("PRIVMSG", func(e *irc.Event) {
					// confからChannelを引っ張る
					fmt.Printf("Message [%+v]\n", e.Arguments)

					for _, c := range conf.Channels {
						if c.Channel == e.Arguments[0] {
							message := fmt.Sprintf("<%s> %s", e.Nick, e.Message())
							SendMessage(c.SlackChannel, message)
							break;
						}
					}
				})
			config.Servers[offset].IRC = conn
			go conn.Loop()
		}()
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
