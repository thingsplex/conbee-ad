package zigbee

import "github.com/futurehomeno/fimpgo"

type FimpToConbeeRouter struct {
	host         string
	apiKey       string
	inboundMsgCh chan []byte
	mqt *fimpgo.MqttTransport
	instanceId string
}



