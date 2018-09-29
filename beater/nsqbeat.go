package beater

import (
	"fmt"
	"time"

	"github.com/json-iterator/go"

	"github.com/mschneider82/nsqbeat/config"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	nsq "github.com/nsqio/go-nsq"
)

// Nsqbeat struct
type Nsqbeat struct {
	done     chan string
	message  chan string
	config   config.Config
	client   beat.Client
	consumer *nsq.Consumer
}

// New creates a beat
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

func (bt *Nsqbeat) createEventWithJSONKeys(jsonData map[string]interface{}) (*beat.Event, error) {
	event := beat.Event{
		Fields: common.MapStr{
			"type": bt.config.Type,
		},
	}

	for key, value := range jsonData {
		// if a json key is @timestamp convert the string to a time object
		// and set this time to Timestamp
		if key == "@timestamp" {
			// currently a correct time format for @timestamp is expected:
			t, err := time.Parse(bt.config.Timelayout, fmt.Sprintf("%v", value))
			if err != nil {
				logp.Err("Error putting: %s", err.Error())
			} else {
				event.Timestamp = t
			}
		} else {
			// if not a timestamp just push the key and value to the event
			_, err := event.PutValue(key, value)
			if err != nil {
				return &event, err
			}
		}
	}

	if event.Timestamp.IsZero() {
		// Set the timestamp to Now() if it wasnt previously set to a value.
		event.Timestamp = time.Now()
	}
	return &event, nil
}

// Run starts nsq consumer and publishs messages to beat
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
		msg:
			switch bt.config.Codec {
			case "json":
				// Doing the Decoding of k&v if codec is json
				var json = jsoniter.ConfigCompatibleWithStandardLibrary
				var jsonData map[string]interface{}

				err := json.Unmarshal([]byte(ev), &jsonData)
				if err != nil {
					logp.Err("Error with Json Data: %v", err)
					break msg
				}

				event, err := bt.createEventWithJSONKeys(jsonData)
				if err != nil {
					logp.Err("Error putting: %s", err.Error())
					break msg
				}

				bt.client.Publish(*event)
				logp.Info("Event send.")

			default:
				// just for Plain text events
				// generate @timestamp by using Now()
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":    bt.config.Type,
						"message": ev,
					},
				}
				bt.client.Publish(event)
				logp.Info("Event send.")
			}
		}
	}
}

// Stop gracefully shutdown
func (bt *Nsqbeat) Stop() {
	bt.consumer.Stop()
	err := bt.client.Close()
	if err != nil {
		logp.Err("Error while stopping: %s", err.Error())
	}
	close(bt.done)
}
