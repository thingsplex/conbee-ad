package zigbee

import (
	log "github.com/Sirupsen/logrus"
	"github.com/futurehomeno/fimpgo"
	"github.com/gorilla/websocket"
	"net/url"
	"strconv"
	"time"
)

type ConbeeEvent struct {
	MsgType      string                 `json:"t"`
	EventType    string                 `json:"e"`
	ResourceType string                 `json:"r"`
	Id           string                 `json:"id"`
	State        map[string]interface{} `json:"state"`
}

type ConbeeToFimpRouter struct {
	host         string
	apiKey       string
	inboundMsgCh chan []byte
	client       *websocket.Conn
	isRunning    bool
	connectionRetryCounter int
	maxConnRetry           int
	mqt *fimpgo.MqttTransport
	instanceId string
}

func NewConbeeToFimpRouter(host string,apiKey string,transport *fimpgo.MqttTransport,instanceId string) *ConbeeToFimpRouter {
	
	return &ConbeeToFimpRouter{host: host,apiKey:apiKey,mqt:transport,instanceId:instanceId}
}



func (cr *ConbeeToFimpRouter) connect() error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<VincClient> Process CRASHED with error : ", r)
		}
	}()
	u := url.URL{Scheme: "ws", Host: cr.host, Path: ""}
	log.Infof("<conb-to-fimp> Connecting to %s", u.String())
	var err error
	cr.client, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("<conb-to-fimp> Dial error", err)
		cr.isRunning = false
		return err
	}
	log.Infof("<conb-to-fimp> Connected")
	return err
}

func (cr *ConbeeToFimpRouter) Start() error {
	err := cr.connect()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("<conb-to-fimp> Process CRASHED with error : ", r)
				cr.isRunning = false
			}
		}()
		defer cr.client.Close()
		for {
			evt := ConbeeEvent{}
			err := cr.client.ReadJSON(&evt)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					log.Error("<conb-to-fimp> CloseError : %v", err)
					if cr.connectionRetryCounter < cr.maxConnRetry {
						log.Info("<conb-to-fimp> Reconnecting after 6 seconds...")
						cr.connectionRetryCounter++
						time.Sleep(time.Second * 6)
						cr.connect()
						continue
					} else {
						cr.connectionRetryCounter = 0
						break
					}
				} else {
					//log.Errorf(" Other error (cmd:%s,comp:%s) : %v",vincMsg.Msg.Data.Cmd,vincMsg.Msg.Data.Component, err)
					continue
				}
			}
			log.Debug("<conb-to-fimp> New event ", evt)
			switch evt.EventType {
			case "added":
				log.Info("<conb-to-fimp> New thing added")
			case "changed":
				cr.mapSensorEvent(&evt)


			}
		}
	}()
	return err
}

func (cr *ConbeeToFimpRouter) SendInclusionReport(addr string) {

}

func (cr *ConbeeToFimpRouter) mapSensorEvent(evt *ConbeeEvent) {
	var msg * fimpgo.FimpMessage
	var adr *fimpgo.Address
	for k := range evt.State {
		if evt.ResourceType == "sensors" {
			log.Debug("state ",k)
			switch k{
			case "open":
				val,ok := evt.State[k].(bool)
				if !ok {
					log.Debug("can't parse the open value")
					continue
				}
				msg = fimpgo.NewBoolMessage("evt.open.report", "sensor_contact", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "zigbee", ResourceAddress: cr.instanceId, ServiceName: "sensor_contact", ServiceAddress: evt.Id}

			case "buttonevent":
				val,ok := evt.State[k].(float64)
				if !ok {
					log.Debug("can't parse the buttonevent value")
					continue
				}
				valS := strconv.FormatFloat(val,'f',-1,32)
				msg = fimpgo.NewStringMessage("evt.scene.report", "scene_ctrl", valS, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "zigbee", ResourceAddress: cr.instanceId, ServiceName: "scene_ctrl", ServiceAddress: evt.Id}

			case "presence":
				val,ok := evt.State[k].(bool)
				if !ok {
					log.Debug("can't parse the presence value")
					continue
				}
				msg = fimpgo.NewBoolMessage("evt.presence.report", "sensor_presence", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "zigbee", ResourceAddress: cr.instanceId, ServiceName: "sensor_presence", ServiceAddress: evt.Id}

			case "temperature":
				val,ok := evt.State[k].(float64)
				if !ok {
					log.Debug("can't parse the temp reading")
					continue
				}
				val = val/100
				msg = fimpgo.NewFloatMessage("evt.sensor.report", "sensor_temp", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "zigbee", ResourceAddress: cr.instanceId, ServiceName: "sensor_temp", ServiceAddress: evt.Id}
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
				msg = fimpgo.NewFloatMessage("evt.sensor.report", "humid_sensor", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "zigbee", ResourceAddress: cr.instanceId, ServiceName: "humid_sensor", ServiceAddress: evt.Id}
			}
		}else if evt.ResourceType == "lights" {
			switch k {
			case "bri":
				val,ok := evt.State[k].(float64)
				if !ok {
					log.Debug("can't parse the bri value")
					continue
				}
				msg = fimpgo.NewIntMessage("evt.lvl.report", "out_lvl_switch", int64(val), nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "zigbee", ResourceAddress: cr.instanceId, ServiceName: "out_lvl_switch", ServiceAddress: evt.Id}

				// level
			case "on":
				val,ok := evt.State[k].(bool)
				if !ok {
					log.Debug("can't parse the on value")
					continue
				}
				msg = fimpgo.NewBoolMessage("evt.binary.report", "out_lvl_switch", val, nil, nil, nil)
				adr = &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "zigbee", ResourceAddress: cr.instanceId, ServiceName: "out_lvl_switch", ServiceAddress: evt.Id}

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
			
		}
		if msg != nil {
			log.Debug("Sending event")
			cr.mqt.Publish(adr,msg)
			msg = nil
		}

	}
			
	
}
