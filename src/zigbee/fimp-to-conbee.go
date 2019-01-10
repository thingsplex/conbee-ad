package zigbee

import (
	"bytes"
	"encoding/json"
	"github.com/alivinco/conbee-ad/model"
	"github.com/futurehomeno/fimpgo"
	log "github.com/Sirupsen/logrus"
	"io"
	"net/http"
)

type FimpToConbeeRouter struct {
	conbeeBaseURL   string
	apiKey       string
	inboundMsgCh fimpgo.MessageCh
	mqt          *fimpgo.MqttTransport
	instanceId   string
	httpClient *http.Client
}

func NewFimpToConbeeRouter(conbeeHost string,mqt *fimpgo.MqttTransport) *FimpToConbeeRouter {
	fc := FimpToConbeeRouter{conbeeBaseURL: conbeeHost, inboundMsgCh: make(fimpgo.MessageCh,5),mqt:mqt}
	fc.mqt.RegisterChannel("ch1",fc.inboundMsgCh)
	fc.httpClient = &http.Client{}
	fc.conbeeBaseURL = "http://legohome.local/api/841CC054BE"
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

	case "out_bin_switch":
		log.Debug("Sending switch")
		val,_ := newMsg.Payload.GetBoolValue()
		req := model.ConnbeeLightRequest{On:val}
		var resp interface{}
		log.Debug("Request ",req)
		respH , err :=  fc.SendConbeeRequest("PUT","lights/2/state",req,resp)
		if err != nil {
			log.Error("Response error ",err)
		}
		log.Debug("Status code = ",respH.StatusCode)

		//
	case "zigbee":
		switch newMsg.Payload.Type {
		case "cmd.network.get_all_nodes":
			// response evt.network.all_nodes_report
		case "cmd.thing.inclusion":
			// open/close network
		case "cmd.thing.remove":
			// remove device from network
		}
		//

	}

}

func (fc *FimpToConbeeRouter) SendConbeeRequest(method, path string, request interface{},response interface{}) (*http.Response, error) {

	var buf io.ReadWriter
	if request != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(request)
		if err != nil {
			return nil, err
		}
	}
	log.Debug("Sending to ",fc.conbeeBaseURL+"/"+path)
	log.Debug("Request ",request)
	req, err := http.NewRequest(method, fc.conbeeBaseURL+"/"+path, buf)
	if err != nil {
		return nil, err
	}
	if request != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	//req.Header.Set("User-Agent", c.UserAgent)
	resp, err := fc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(response)
	return resp, err

}
