package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alivinco/conbee-ad/model"
	"github.com/alivinco/conbee-ad/zigbee"
	"github.com/futurehomeno/fimpgo"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
)

func SetupLog(logfile string, level string, logFormat string) {
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.999"})
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, ForceColors: true, TimestampFormat: "2006-01-02T15:04:05.999"})
	}

	logLevel, err := log.ParseLevel(level)
	if err == nil {
		log.SetLevel(logLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	if logfile != "" {
		l := lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    5, // megabytes
			MaxBackups: 2,
		}
		log.SetOutput(&l)
	}

}

func main() {
	configs := model.Configs{}
	var configFile string
	flag.StringVar(&configFile, "c", "", "Config file")
	flag.Parse()
	if configFile == "" {
		configFile = "./config.json"
	} else {
		fmt.Println("Loading configs from file ", configFile)
	}
	configFileBody, err := ioutil.ReadFile(configFile)
	err = json.Unmarshal(configFileBody, &configs)
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting ThingsPlexServiceTemplate----------------")

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI,configs.MqttClientIdPrefix,configs.MqttUsername,configs.MqttPassword,true,1,1)
	err = mqtt.Start()
	if err != nil {
		log.Error("Failed to connect ot broker. Error:",err.Error())
	}else {
		log.Info("Connected")
	}


	conFimpRouter := zigbee.NewConbeeToFimpRouter("legohome.local:443","841CC054BE",mqtt,configs.InstanceAddress)
	conFimpRouter.Start()

	select {

	}
}
