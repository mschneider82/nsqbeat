package beater

import (
	"fmt"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/mschneider82/nsqbeat/config"
)

func TestNsqbeat_createEventWithJSONKeys(t *testing.T) {
	type args struct {
		jsonData map[string]interface{}
	}
	bt := &Nsqbeat{
		done:    make(chan string),
		message: make(chan string),
		config:  config.DefaultConfig,
	}

	jsondata := `{"testkey": "value", "@timestamp": "2018-09-29T13:05:01.001Z"}`
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var jsonData map[string]interface{}
	json.Unmarshal([]byte(jsondata), &jsonData)

	tests := []struct {
		name string
		bt   *Nsqbeat
		args args
	}{
		{"test1", bt, args{jsonData}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bt.createEventWithJSONKeys(tt.args.jsonData)
			if err != nil {
				t.Errorf("Nsqbeat.createEventWithJSONKeys() error = %v", err)
				return
			}

			gotTSString := fmt.Sprintf("%v", got.Timestamp)
			if gotTSString != "2018-09-29 13:05:01.001 +0000 UTC" {
				t.Errorf("Nsqbeat.createEventWithJSONKeys() = %v, want %v", gotTSString, "2018-09-29 13:05:01.001 +0000 UTC")
			}
		})
	}
}
