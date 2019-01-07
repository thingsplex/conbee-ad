package zigbee

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/url"
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
}

func NewConbeeToFimpRouter(host string,apiKey string) *ConbeeToFimpRouter {
	return &ConbeeToFimpRouter{host: host,apiKey:apiKey}
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
			switch evt.MsgType {
			case "added":
				log.Info("<conb-to-fimp> New thing added")
			case "changed":


			}
		}
	}()
	return err
}

func (cr *ConbeeToFimpRouter) SendInclusionReport(addr string) {
	
}
