// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	LookupDaemons []string `config:"lookupdhttpaddrs"`
	Topic         string   `config:"topic"`
	Channel       string   `config:"channel"`
	MaxInFlight   int      `config:"maxinflight"`
	Codec         string   `config:"codec"`
	Type          string   `config:"type"`
	Timelayout    string   `config:"timelayout"`
}

var DefaultConfig = Config{
	LookupDaemons: []string{"127.0.0.1:4161"},
	Topic:         "topicname",
	Channel:       "channelname",
	MaxInFlight:   200,
	Codec:         "json",
	Type:          "nsqbeat",
	Timelayout:    "2006-01-02T15:04:05.000Z",
}
