package main

import (
	"flag"
	"fmt"
	"github.com/alivinco/conbee-ad/conbee"
	"github.com/alivinco/conbee-ad/model"
	"github.com/alivinco/conbee-ad/zigbee"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	"github.com/futurehomeno/fimpgo/fimptype"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
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

func getDiscoveryResource() discovery.Resource {
	adInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.network.get_all_nodes",
		ValueType: "null",
		Version:   "1",
	},{
		Type:      "in",
		MsgType:   "cmd.thing.get_inclusion_report",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.thing.inclusion",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.thing.delete",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.thing.inclusion_report",
		ValueType: "object",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.thing.exclusion_report",
		ValueType: "object",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.network.all_nodes_report",
		ValueType: "object",
		Version:   "1",
	}}


	adService := fimptype.Service{
		Name:    "conbee",
		Alias:   "Network managment",
		Address: "/rt:ad/rn:conbee/ad:1",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       adInterfaces,
	}
	return  discovery.Resource{
		ResourceName:"conbee",
		ResourceType:discovery.ResourceTypeAd,
		Author:"aleksandrs.livincovs@gmail.com",
		IsInstanceConfigurable:false,
		InstanceId:"1",
		Version:"1",
		AdapterInfo:discovery.AdapterInfo{
			Technology:"conbee",
			FwVersion:"all",
			NetworkManagementType:"inclusion_exclusion",
			Services:[]fimptype.Service{adService},
		},
	}

}

func main() {

	var configFile string
	flag.StringVar(&configFile, "c", "", "Config file")
	flag.Parse()
	if configFile == "" {
		configFile = "./config.json"
	} else {
		fmt.Println("Loading configs from file ", configFile)
	}
	configs := model.NewConfigs(configFile)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("-------------- Starting conbee-ad ----------------")

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI,configs.MqttClientIdPrefix,configs.MqttUsername,configs.MqttPassword,true,1,1)
	err = mqtt.Start()
	if err != nil {
		log.Error("Failed to connect ot broker. Error:",err.Error())
	}else {
		log.Info("Connected")
	}
	mqtt.Subscribe("pt:j1/+/rt:dev/rn:conbee/ad:1/#")
	mqtt.Subscribe("pt:j1/+/rt:ad/rn:conbee/ad:1")

	//"841CC054BE"
	// "legohome.local:443"


	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(getDiscoveryResource())
	responder.Start()

	conbeeClient := conbee.NewClient(configs.ConbeeUrl)
	conbeeClient.SetApiKey("841CC054BE")

	netService := zigbee.NewNetworkService(mqtt,conbeeClient)
	conFimpRouter := zigbee.NewConbeeToFimpRouter(mqtt, conbeeClient,netService,configs.InstanceAddress)
	conFimpRouter.Start()
	fimpRouter := zigbee.NewFimpToConbeeRouter(mqtt,conbeeClient,netService)
	fimpRouter.Start()

	select {

	}
}
