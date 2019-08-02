package zigbee

import (
	log "github.com/sirupsen/logrus"
	"github.com/alivinco/conbee-ad/conbee"
	"github.com/futurehomeno/fimpgo"
	"github.com/gorilla/websocket"
	"strconv"
)

type ConbeeToFimpRouter struct {
	host         string
	apiKey       string
	inboundMsgCh chan []byte
	client       *websocket.Conn
	isRunning    bool
	mqt *fimpgo.MqttTransport
	instanceId string
	conbeeClient *conbee.Client
	conbeeEventStream chan conbee.ConbeeEvent
	netService *NetworkService
	batteryLevels map[string]int
}

func NewConbeeToFimpRouter(transport *fimpgo.MqttTransport,conbeeClient *conbee.Client,netService *NetworkService,instanceId string) *ConbeeToFimpRouter {
	
	return &ConbeeToFimpRouter{conbeeClient:conbeeClient,mqt:transport,instanceId:instanceId,netService:netService}
}

func (cr *ConbeeToFimpRouter) Start() {
	cr.batteryLevels = map[string]int{}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("<conb-to-fimp> Process CRASHED with error : ", r)
				cr.isRunning = false
			}
		}()
		var err error
		cr.conbeeEventStream ,err = cr.conbeeClient.GetMsgStream()
		if err != nil {
			log.Error("Cant get conbee event stream . Err :",err)
			return
		}
		for {
			log.Debug("<conb-to-fimp> Waiting for new event")
			newEvent := <- cr.conbeeEventStream
			switch newEvent.EventType {
			case "added":
				log.Info("<conb-to-fimp> New thing added")
				cr.netService.SendInclusionReport(newEvent.ResourceType,newEvent.Id)
			case "deleted":
				log.Info("<conb-to-fimp> New thing added")
				cr.netService.DeleteThing(newEvent.Id)
			case "changed":
				cr.mapSensorEvent(&newEvent)
			default:
				log.Debug("MsgType :",newEvent.MsgType)
			}
		}
	}()

}

//func (cr *ConbeeToFimpRouter) isBatteryLevelChanged(event *conbee.ConbeeEvent) bool {
//	 oldValue , ok := cr.batteryLevels[event.Id]
//	 if ok {
//		 if oldValue == event.Sensor.Config.Battery {
//		 	return false
//		 }else {
//			 cr.batteryLevels[event.Id] = event.Sensor.Config.Battery
//		 	return true
//		 }
//	 }else {
//		cr.batteryLevels[event.Id] = event.Sensor.Config.Battery
//	 	return true
//	 }
//}

func (cr *ConbeeToFimpRouter) mapSensorEvent(evt *conbee.ConbeeEvent) {
	var msg * fimpgo.FimpMessage
	var adr *fimpgo.Address
	var serviceAddress string
	if evt.Sensor.Config.Battery !=nil {
		msg = fimpgo.NewIntMessage("evt.lvl.report", "battery",int64(*evt.Sensor.Config.Battery), nil, nil, nil)
		adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "battery", ServiceAddress: evt.Id}
		cr.mqt.Publish(adr,msg)
	}

	for k := range evt.State {
		if evt.ResourceType == "sensors" {
			serviceAddress = "s"+evt.Id+"_0"
			log.Debug("state ",k)
			switch k{
			case "open":
				val,ok := evt.State[k].(bool)
				if !ok {
					log.Debug("can't parse the open value")
					continue
				}
				msg = fimpgo.NewBoolMessage("evt.open.report", "sensor_contact", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "sensor_contact", ServiceAddress: serviceAddress}

			case "buttonevent":
				val,ok := evt.State[k].(float64)
				if !ok {
					log.Debug("can't parse the buttonevent value")
					continue
				}
				valS := strconv.FormatFloat(val,'f',-1,32)
				msg = fimpgo.NewStringMessage("evt.scene.report", "scene_ctrl", valS, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "scene_ctrl", ServiceAddress: serviceAddress}

			case "presence":
				val,ok := evt.State[k].(bool)
				if !ok {
					log.Debug("can't parse the presence value")
					continue
				}
				msg = fimpgo.NewBoolMessage("evt.presence.report", "sensor_presence", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "sensor_presence", ServiceAddress: serviceAddress}

			case "temperature":
				val,ok := evt.State[k].(float64)
				if !ok {
					log.Debug("can't parse the temp reading")
					continue
				}
				val = val/100
				msg = fimpgo.NewFloatMessage("evt.sensor.report", "sensor_temp", val, map[string]string{"unit":"C"}, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "sensor_temp", ServiceAddress: serviceAddress}
				// sensor temp
			case "flag":
				// ?
			case "status":
				//
			case "humidity":
				val,ok := evt.State[k].(float64)
				if !ok {
					log.Debug("can't parse the temp reading")
					continue
				}
				val = val/100
				msg = fimpgo.NewFloatMessage("evt.sensor.report", "sensor_humid", val,  map[string]string{"unit":"%"}, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "sensor_humid", ServiceAddress: serviceAddress}
			}
		}else if evt.ResourceType == "lights" {
			serviceAddress = "l"+evt.Id+"_0"
			switch k {
			case "bri":
				val,ok := evt.State[k].(float64)
				if !ok {
					log.Debug("can't parse the bri value")
					continue
				}
				msg = fimpgo.NewIntMessage("evt.lvl.report", "out_lvl_switch", int64(val), nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "out_lvl_switch", ServiceAddress: serviceAddress}

				// level
			case "on":
				val,ok := evt.State[k].(bool)
				if !ok {
					log.Debug("can't parse the on value")
					continue
				}
				msg = fimpgo.NewBoolMessage("evt.binary.report", "out_lvl_switch", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "conbee", ResourceAddress: cr.instanceId, ServiceName: "out_lvl_switch", ServiceAddress: serviceAddress}

				//
			case "hue":
				//
			case "sat":
				//
			case "ct":
				//
			case "xy":
				//
			case "alert":
				//
			case "effect":
				//
			case "colorloopspeed":
				//
			case "transitiontime":
				//
				
			}	
		}else if evt.ResourceType == "groups" {
			
		} else {
			log.Debug("Unknown Resouce type :",evt.ResourceType)
		}

		if msg != nil {
			log.Debug("Sending event")
			cr.mqt.Publish(adr,msg)
			msg = nil
		}

	}
			
	
}
