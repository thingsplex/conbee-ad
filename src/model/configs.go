package model

import (
	"encoding/json"
	"fmt"
	"github.com/futurehomeno/fimpgo/edgeapp"
	"github.com/futurehomeno/fimpgo/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const ServiceName  = "conbee"

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
	WorkDir               string `json:"-"`
	ConfiguredAt          string `json:"configured_at"`
	ConfiguredBy          string `json:"configured_by"`
}


func NewConfigs(workDir string) *Configs {
	conf := &Configs{WorkDir: workDir}
	conf.path = filepath.Join(workDir,"data","config.json")
	if !utils.FileExists(conf.path) {
		log.Info("Config file doesn't exist.Loading default config")
		defaultConfigFile := filepath.Join(workDir,"defaults","config.json")
		err := utils.CopyFile(defaultConfigFile,conf.path)
		if err != nil {
			fmt.Print(err)
			panic("Can't copy config file.")
		}
	}
	return conf
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

func (cf *Configs) GetDataDir()string {
	return filepath.Join(cf.WorkDir,"data")
}

func (cf *Configs) GetDefaultDir()string {
	return filepath.Join(cf.WorkDir,"defaults")
}

func (cf * Configs) LoadDefaults()error {
	configFile := filepath.Join(cf.WorkDir,"data","config.json")
	os.Remove(configFile)
	log.Info("Config file doesn't exist.Loading default config")
	defaultConfigFile := filepath.Join(cf.WorkDir,"defaults","config.json")
	return utils.CopyFile(defaultConfigFile,configFile)
}

func (cf *Configs) InitDefault() {
	cf.InstanceAddress = "1"
	cf.MqttServerURI = "tcp://localhost:1883"
	cf.MqttClientIdPrefix = "conbee-ad"
	cf.LogFile = "/var/log/thingsplex/conbee-ad/conbee-ad.log"
	cf.LogLevel = "info"
	cf.LogFormat = "text"
}

func (cf *Configs) IsConfigured()bool {
	if cf.ConbeeUrl == "" || cf.ConbeeApiKey == "" {
		return false
	}
	return true
}

type ConfigReport struct {
	OpStatus string `json:"op_status"`
	AppState edgeapp.AppStates `json:"app_state"`
}