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
	"strings"
	"sync"
	"time"
)

type Config struct {
	Apiversion          string `json:"apiversion"`
	Bridgeid            string `json:"bridgeid"`
	Devicename          string `json:"devicename"`
	Dhcp                bool   `json:"dhcp"`
	Fwversion           string `json:"fwversion"`
	Gateway             string `json:"gateway"`
	Ipaddress           string `json:"ipaddress"`
	Linkbutton          bool   `json:"linkbutton"`
	Localtime           string `json:"localtime"`
	Mac                 string `json:"mac"`
	Modelid             string `json:"modelid"`
	Name                string `json:"name"`
	Netmask             string `json:"netmask"`
	Networkopenduration int    `json:"networkopenduration"`
	Ntp                 string `json:"ntp"`
	Panid               int    `json:"panid"`
	Portalservices      bool   `json:"portalservices"`
	Proxyaddress        string `json:"proxyaddress"`
	Proxyport           int    `json:"proxyport"`
	Rfconnected         bool   `json:"rfconnected"`
	Swupdate            struct {
		Notify      bool   `json:"notify"`
		Text        string `json:"text"`
		Updatestate int    `json:"updatestate"`
		URL         string `json:"url"`
	} `json:"swupdate"`
	Swversion          string `json:"swversion"`
	Timeformat         string `json:"timeformat"`
	Timezone           string `json:"timezone"`
	UTC                string `json:"UTC"`
	UUID               string `json:"uuid"`
	Websocketnotifyall bool   `json:"websocketnotifyall"`
	Websocketport      int    `json:"websocketport"`
	Whitelist          struct {
	} `json:"whitelist"`
	Zigbeechannel int `json:"zigbeechannel"`
}

type Client struct {
	host                 string
	apiKey               string
	wsClient             *websocket.Conn
	httpClient           *http.Client
	isWsConnectionActive bool
	msgStream            chan ConbeeEvent
	connectionRetryCounter int
	maxConnRetry           int
	fullState            FullState
	stateMux             sync.RWMutex
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
	cb.stateMux = sync.RWMutex{}
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
	port := 443
	conf , err := rc.GetConbeeConfigs()
	if err != nil {
		log.Error("<conb-client> Failed to load config. Err:",err.Error())
	}else {
		if conf.Websocketport != 0 {
			port = conf.Websocketport
		}
	}

	hostS := strings.Split(rc.host,":")
	host := fmt.Sprintf("%s:%d",hostS[0],port)
	log.Info("<conb-client> Establishing WS connection to host : ",host)
	u := url.URL{Scheme: "ws", Host: host, Path: ""}
	log.Infof("<conb-client> Connecting to %s", u.String())
	rc.wsClient, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("<conb-client> Dial error", err)
		rc.isWsConnectionActive = false
		return  err
	}
	rc.isWsConnectionActive = true
	rc.LoadFullState()
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

func (rc *Client) GetConbeeConfigs() (*Config,error) {
	config := Config {}
	_, err := rc.SendConbeeRequest("GET", "config", nil, &config)
	if err != nil {
		log.Error("Can't get device descriptor . Err :", err)
		return nil ,err
	}
	return &config,err
}

func (rc *Client) LoadFullState() error {
	_, err := rc.SendConbeeRequest("GET", "", nil, &rc.fullState)
	if err != nil {
		log.Error("Can't load full state . Err :", err)
	}else {
		log.Debug("Full state loaded successfully.")
	}

	return err
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

func (rc *Client) GetLightById(id string)Light {
	rc.stateMux.RLock()
	defer rc.stateMux.RUnlock()
	if rc.fullState.Lights == nil {
		return Light{}
	}
	l , _ := rc.fullState.Lights[id]
	return l
}

func (rc *Client) GetSensorById(id string)Sensor {
	rc.stateMux.RLock()
	defer rc.stateMux.RUnlock()
	if rc.fullState.Sensors == nil {
		return Sensor{}
	}
	l , _ := rc.fullState.Sensors[id]
	return l

}
