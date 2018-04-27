package beater

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mschneider82/nsqbeat/config"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	nsq "github.com/nsqio/go-nsq"
)

type Nsqbeat struct {
	done     chan string
	message  chan string
	config   config.Config
	client   beat.Client
	consumer *nsq.Consumer
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	nsqConfig := nsq.NewConfig()

	consumer, err := nsq.NewConsumer(c.Topic, c.Channel, nsqConfig)
	if err != nil {
		return nil, fmt.Errorf("Error starting consumer %s", err)
	}

	consumer.ChangeMaxInFlight(c.MaxInFlight)

	bt := &Nsqbeat{
		done:     make(chan string),
		message:  make(chan string),
		config:   c,
		consumer: consumer,
	}

	return bt, nil
}

func (bt *Nsqbeat) Run(b *beat.Beat) error {
	logp.Info("nsqbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	bt.consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		logp.Info("Got a message: %s", message.Body)
		bt.message <- string(message.Body)
		return nil
	}))

	if err := bt.consumer.ConnectToNSQLookupds(bt.config.LookupDaemons); err != nil {
		return fmt.Errorf("Error connecting to lookupds %s", err)
	}

	for {
		select {
		case <-bt.done:
			return nil
		case ev := <-bt.message:
			// Doing the Decoding of k&v if codec is json
			if bt.config.Codec == "json" {
				var jsonData map[string]interface{}

				event := beat.Event{
					Fields: common.MapStr{
						"type": bt.config.Type,
					},
				}

				err := json.Unmarshal([]byte(ev), &jsonData)
				if err != nil {
					logp.Err("Error with Json Data: %v", err)
				} else {
					for key, value := range jsonData {
						// if a json key is @timestamp convert the string to a time object
						// and set this time to Timestamp
						if key == "@timestamp" {
							// currently a correct time format for @timestamp is expected:
							layout := "2006-01-02T15:04:05.000Z"
							var t time.Time
							t, err = time.Parse(layout, fmt.Sprintf("%v", value))
							if err != nil {
								logp.Err("Error putting: ", err)
							} else {
								event.Timestamp = t
							}
						} else {
							// if not a timestamp just push the key and value to the event
							_, err = event.PutValue(key, value)
							if err != nil {
								logp.Err("Error putting: ", err)
							}
						}
					}
					if err == nil {
						if event.Timestamp.IsZero() {
							// Set the timestamp to Now() if it wasnt previously set to a value.
							event.Timestamp = time.Now()
						}
						bt.client.Publish(event)
						logp.Info("Event send.")
					}
				}
			} else {
				// just for Plain text events
				// generate @timestamp by using Now()
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":    bt.config.Type,
						"message": string(ev),
					},
				}
				bt.client.Publish(event)
				logp.Info("Event send.")
			}
		}
	}
}

func (bt *Nsqbeat) Stop() {
	bt.consumer.Stop()
	bt.client.Close()
	close(bt.done)
}
