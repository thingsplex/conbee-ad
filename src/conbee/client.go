package conbee

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
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

func (rc *Client) SetApiKeyAndHost(apiKey string, host string) {
	rc.apiKey = apiKey
	rc.host = host
}

func NewClient(conbeeBaseURL string) *Client {
	cb := &Client{host: conbeeBaseURL}
	//"http://legohome.local/api/841CC054BE"
	cb.httpClient = &http.Client{Timeout: 15 * time.Second}
	cb.msgStream = make(chan ConbeeEvent,10)
	cb.maxConnRetry = 2000
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

func (rc *Client) Login(username, password string) error {
	if username == "" {
		username = "delight"
	}
	httpClient := &http.Client{}
	credentials := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",username,password)))
	payloadM := map[string]string{"devicetype":"conbee-ad","login":username}
	payloadB , _ := json.Marshal(payloadM)
	path := "http://"+rc.host + "/api"
	log.Debug("Sending to ", path)
	log.Debug("Credentials ", credentials)
	req, err := http.NewRequest("POST", path, bytes.NewBuffer(payloadB))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		//b,_ := ioutil.ReadAll(resp.Body)
		//log.Debug("Response code :",resp.Status)
		log.Error("Error :",err.Error())
		return err
	}

	bBody,err := ioutil.ReadAll(resp.Body)
	log.Debug("Status :",resp.Status)
	log.Debug("Response :",string(bBody))

	var response []map[string]map[string]string

	if resp != nil {
		err = json.Unmarshal(bBody,&response)
		if err != nil {
			log.Error("Login response is in wrong format",err)
			return err
		}
	}

	if len(response)>0 {
		uo , ok := response[0]["success"]
		if ok {
			rc.apiKey,_ = uo["username"]
			log.Debug("Api key:",rc.apiKey)
		}
	}
	return nil
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
