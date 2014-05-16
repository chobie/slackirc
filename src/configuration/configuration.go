package configuration

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"fmt"
	"github.com/thoj/go-ircevent"
)

type Channel struct {
	Channel string `toml:"channel"`
	SlackChannel string `toml:"slack_channel"`
	SlackChannelToken string `toml:"slack_channel_token"`
	ReadOnly bool `toml:"readonly"`
}

type Http struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

type Server struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
	Nick string `toml:"nick"`
	Name string `toml:"name"`
	Pass string `toml:"pass"`
	NickServePass string `toml:"nick_serve_pass"`
	Channels []Channel `toml:"channels"`
	IRC *irc.Connection
}

type Slack struct {
	Endpoint string `toml:"endpoint"`
	Token string `toml:"token"`
	AcceptNicks []string `toml:"accept_nicks"`
}

type Config struct {
	Slack Slack `toml:"slack"`
	Http Http`toml:"http"`
	Servers []Server `toml:"servers"`
}

func LoadConfiguration(configFile string) (*Config, error) {
	config := &Config{}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	if _, err2 := toml.Decode(string(data), config); err != nil {
		fmt.Printf("string: %s\n", string(data))
		return config, err2
	}

	return config, nil
}
