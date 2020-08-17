package main

import (
	"flag"
	"fmt"
	"github.com/alivinco/conbee-ad/conbee"
	"github.com/alivinco/conbee-ad/model"
	"github.com/alivinco/conbee-ad/utils"
	"github.com/alivinco/conbee-ad/zigbee"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
)

func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := model.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}
	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("-------------- Starting conbee-ad ----------------")

	appLifecycle.SetAppState(edgeapp.AppStateStarting, nil)
	appLifecycle.SetConnectionState(edgeapp.ConnStateDisconnected)
	appLifecycle.SetAuthState(edgeapp.AuthStateNotAuthenticated)
	appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)

	if configs.IsConfigured() {
		appLifecycle.SetAppState(edgeapp.AppStateRunning, nil)
		appLifecycle.SetConfigState(edgeapp.ConfigStateConfigured)
		appLifecycle.SetAuthState(edgeapp.AuthStateAuthenticated)
	}else {
		appLifecycle.SetAppState(edgeapp.AppStateNotConfigured, nil)
		log.Info("<main> Application is not configured.Waiting for configurations ")
	}

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	if err != nil {
		log.Error("Failed to connect ot broker. Error:", err.Error())
	} else {
		log.Info("Connected")
	}

	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()

	//"841CC054BE"
	// "legohome.local:443"
	conbeeClient := conbee.NewClient(configs.ConbeeUrl)
	netService := zigbee.NewNetworkService(mqtt, conbeeClient)
	fimpRouter := zigbee.NewFimpToConbeeRouter(mqtt, conbeeClient, netService ,appLifecycle,configs)
	conFimpRouter := zigbee.NewConbeeToFimpRouter(mqtt, conbeeClient, netService, configs.InstanceAddress)

	fimpRouter.Start()

	for {
		appLifecycle.WaitForState("main", model.AppStateRunning)
		conbeeClient.SetApiKeyAndHost(configs.ConbeeApiKey, configs.ConbeeUrl)
		if err := conFimpRouter.Start(); err !=nil {
			appLifecycle.PublishEvent(model.EventConfigError,"main",nil)
		}else {
			appLifecycle.SetConnectionState(edgeapp.ConnStateConnected)
			appLifecycle.WaitForState("main",model.AppStateNotConfigured)
		}
	}
}
