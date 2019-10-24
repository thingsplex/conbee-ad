package model

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type Configs struct {
	path                  string
	InstanceAddress       string `json:"instance_address"`
	MqttServerURI         string `json:"mqtt_server_uri"`
	MqttUsername          string `json:"mqtt_server_username"`
	MqttPassword          string `json:"mqtt_server_password"`
	MqttClientIdPrefix    string `json:"mqtt_client_id_prefix"`
	ConbeeUrl             string `json:"conbee_url"`
	ConbeeApiKey          string `json:"conbee_api_key"`
	LogFile               string `json:"log_file"`
	LogLevel              string `json:"log_level"`
	LogFormat             string `json:"log_format"`
	ConfiguredAt          string `json:"configured_at"`
	ConfiguredBy          string `json:"configured_by"`
}

func NewConfigs(path string) *Configs {
	return &Configs{path:path}
}

func (cf * Configs) LoadFromFile() error {
	configFileBody, err := ioutil.ReadFile(cf.path)
	if err != nil {
		cf.InitDefault()
		return cf.SaveToFile()
	}
	err = json.Unmarshal(configFileBody, cf)
	if err != nil {
		return err
	}
	return nil
}

func (cf *Configs) SaveToFile() error {
	cf.ConfiguredBy = "auto"
	cf.ConfiguredAt = time.Now().Format(time.RFC3339)
	bpayload, err := json.Marshal(cf)
	err = ioutil.WriteFile(cf.path, bpayload, 0664)
	if err != nil {
		return err
	}
	return err
}

func (cf *Configs) InitDefault() {
	cf.InstanceAddress = "1"
	cf.MqttServerURI = "tcp://localhost:1883"
	cf.MqttClientIdPrefix = "conbee-ad"
	cf.LogFile = "/var/log/thingsplex/conbee-ad/conbee-ad.log"
	cf.LogLevel = "info"
	cf.LogFormat = "text"
}