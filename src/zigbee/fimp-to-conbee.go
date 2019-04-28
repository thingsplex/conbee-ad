package zigbee

import (
	log "github.com/Sirupsen/logrus"
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
	switch newMsg.Payload.Service {
	case "out_lvl_switch" :
		addr := strings.Replace(newMsg.Addr.ServiceAddress,"_0","",1)
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
		addr := strings.Replace(newMsg.Addr.ServiceAddress,"_0","",1)
		respH , err :=  fc.conbeeClient.SendConbeeRequest("PUT","lights/"+addr+"/state",req,resp)
		if err != nil {
			log.Error("Response error ",err)
		}
		log.Debug("Status code = ",respH.StatusCode)

		//
	case "zigbee":
		switch newMsg.Payload.Type {
		case "cmd.network.get_all_nodes":
			// response evt.network.all_nodes_report
			fc.netService.SendListOfDevices()
		case "cmd.thing.get_inclusion_report":
			nodeId , _ := newMsg.Payload.GetStringValue()
			fc.netService.SendInclusionReport("",nodeId)
		case "cmd.thing.inclusion":
			// open/close network
		case "cmd.thing.delete":
			// remove device from network
		}
		//

	}

}


