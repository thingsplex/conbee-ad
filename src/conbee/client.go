package conbee

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	host                 string
	apiKey               string
	wsClient             *websocket.Conn
	httpClient           *http.Client
	isWsConnectionActive bool
	msgStream            chan ConbeeEvent
	connectionRetryCounter int
	maxConnRetry           int
}

func (rc *Client) ApiKey() string {
	return rc.apiKey
}

func (rc *Client) SetApiKey(apiKey string) {
	rc.apiKey = apiKey
}

func NewClient(conbeeBaseURL string) *Client {
	cb := &Client{host: conbeeBaseURL}
	//"http://legohome.local/api/841CC054BE"
	cb.httpClient = &http.Client{Timeout: 15 * time.Second}
	cb.msgStream = make(chan ConbeeEvent,10)
	cb.maxConnRetry = 100
	return cb
}

func (rc *Client) GetMsgStream() (chan ConbeeEvent, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<VincClient> Process CRASHED with error : ", r)
		} else {

		}
	}()
	if rc.isWsConnectionActive {
		return rc.msgStream , nil
	}
	err := rc.wsConnect()
	if err != nil {
		return nil , err
	}
	rc.startWsEventLoop()
	return rc.msgStream, err
}

func (rc *Client) wsConnect()error {
	u := url.URL{Scheme: "ws", Host: rc.host+":443", Path: ""}
	log.Infof("<conb-client> Connecting to %s", u.String())
	var err error
	rc.wsClient, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("<conb-client> Dial error", err)
		rc.isWsConnectionActive = false
		return  err
	}
	rc.isWsConnectionActive = true
	log.Info("<conb-client> Connected ")
	return nil
}

func (rc *Client) startWsEventLoop() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("<conb-client> Process CRASHED with error : ", r)
			}
			rc.isWsConnectionActive= false
			log.Infof("<conb-client> Event loop was is stopped")
		}()
		log.Debug("<conb-client> Starting event loop . Client is connected = ",rc.isWsConnectionActive)
		var err error
		for {
			evt := ConbeeEvent{}
			if rc.isWsConnectionActive {
				err = rc.wsClient.ReadJSON(&evt)
			}
			if err != nil || !rc.isWsConnectionActive {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) || !rc.isWsConnectionActive {
					rc.isWsConnectionActive = false
					log.Errorf("<conb-client> CloseError : %v", err)
					if rc.connectionRetryCounter < rc.maxConnRetry {
						log.Info("<conb-client> Reconnecting after 10 seconds...")
						rc.connectionRetryCounter++
						time.Sleep(time.Second * 5)
						rc.wsConnect()
						continue
					} else {
						rc.connectionRetryCounter = 0
						break
					}
				} else {
					rc.isWsConnectionActive = false
					log.Errorf("<conb-client> Other error  %v", err)
					time.Sleep(time.Second * 30)
					rc.wsConnect()
					continue
				}
			}
			log.Debug("<conb-client> New event. sending to ch ")
			rc.msgStream <- evt
		}
	}()
}

func Login(username, password string) {

}

func (rc *Client) SendConbeeRequest(method, path string, request interface{}, response interface{}) (*http.Response, error) {
	httpClient := &http.Client{}
	var buf io.ReadWriter
	if request != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(request)
		if err != nil {
			return nil, err
		}
	}
	path = "http://"+rc.host + "/api/" + rc.apiKey + "/" + path
	log.Debug("Sending to ", path)
	log.Debug("Request ", request)
	req, err := http.NewRequest(method, path, buf)
	if err != nil {
		return nil, err
	}
	if request != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	//req.Header.Set("User-Agent", c.UserAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if response != nil {
		err = json.NewDecoder(resp.Body).Decode(response)
	}
	return resp, err

}
