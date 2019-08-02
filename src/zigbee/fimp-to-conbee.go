package zigbee

import (
	log "github.com/sirupsen/logrus"
	"github.com/alivinco/conbee-ad/conbee"
	"github.com/futurehomeno/fimpgo"
	"strings"
)

type FimpToConbeeRouter struct {
	inboundMsgCh fimpgo.MessageCh
	mqt          *fimpgo.MqttTransport
	instanceId   string
	conbeeClient *conbee.Client
	netService *NetworkService
}

func NewFimpToConbeeRouter(mqt *fimpgo.MqttTransport, conbeeClient *conbee.Client,netService *NetworkService) *FimpToConbeeRouter {
	fc := FimpToConbeeRouter{inboundMsgCh: make(fimpgo.MessageCh,5),mqt:mqt,netService:netService}
	fc.mqt.RegisterChannel("ch1",fc.inboundMsgCh)
	fc.conbeeClient = conbeeClient
		return &fc
}

func (fc *FimpToConbeeRouter) Start() {
	go func(msgChan fimpgo.MessageCh) {
		for  {
			select {
			case newMsg :=<- msgChan:
				fc.routeFimpMessage(newMsg)

			}
		}

	}(fc.inboundMsgCh)
}

func (fc *FimpToConbeeRouter) routeFimpMessage(newMsg *fimpgo.Message) {
	log.Debug("New fimp msg")
	addr := strings.Replace(newMsg.Addr.ServiceAddress,"_0","",1)
	switch newMsg.Payload.Service {
	case "out_lvl_switch" :
		addr = strings.Replace(addr,"l","",1)
		switch newMsg.Payload.Type {
		case "cmd.binary.set":
			val,_ := newMsg.Payload.GetBoolValue()
			req := conbee.ConnbeeLightRequest{On:val}
			var resp interface{}
			log.Debug("Request ",req)
			_ , err :=  fc.conbeeClient.SendConbeeRequest("PUT","lights/"+addr+"/state",req,resp)
			if err != nil {
				log.Error("Response error ",err)
			}
			//log.Debug("Status code = ",respH.StatusCode)
		case "cmd.lvl.set":
			val,_ := newMsg.Payload.GetIntValue()
			// 255 - 100%
			// A   - x%
			//x = A * 100 / 255

			req := conbee.ConnbeeLightRequest{Bri:int(val),On:true}
			var resp interface{}
			log.Debug("Request ",req)
			_ , err :=  fc.conbeeClient.SendConbeeRequest("PUT","lights/"+addr+"/state",req,resp)
			if err != nil {
				log.Error("Response error ",err)
			}
		}

	case "out_bin_switch":
		log.Debug("Sending switch")
		val,_ := newMsg.Payload.GetBoolValue()
		req := conbee.ConnbeeLightRequest{On:val}
		var resp interface{}
		log.Debug("Request ",req)
		addr = strings.Replace(addr,"l","",1)
		respH , err :=  fc.conbeeClient.SendConbeeRequest("PUT","lights/"+addr+"/state",req,resp)
		if err != nil {
			log.Error("Response error ",err)
		}
		log.Debug("Status code = ",respH.StatusCode)

		//
	case "conbee":
		switch newMsg.Payload.Type {
		case "cmd.network.get_all_nodes":
			fc.netService.SendListOfDevices()
		case "cmd.thing.get_inclusion_report":
			nodeId , _ := newMsg.Payload.GetStringValue()
			var deviceType string
			if strings.Contains(nodeId,"s") {
				deviceType = "sensors"
				nodeId = strings.Replace(nodeId,"s","",1)
			}else {
				deviceType = "lights"
				nodeId = strings.Replace(nodeId,"l","",1)
			}

			fc.netService.SendInclusionReport(deviceType,nodeId)
		case "cmd.thing.inclusion":
			flag , _ := newMsg.Payload.GetBoolValue()
			fc.netService.OpenNetwork(flag)
		case "cmd.thing.delete":
			// remove device from network
			val,err := newMsg.Payload.GetStrMapValue()
			if err != nil {
				log.Error("Wrong msg format")
				return
			}
			deviceId , ok := val["address"]
			if ok {
				fc.netService.DeleteThing(deviceId)
			}else {
				log.Error("Incorrect address")

			}

		}
		//

	}

}


