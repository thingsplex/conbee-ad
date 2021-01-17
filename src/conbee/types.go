package conbee

const (
	DeviceTypeWindowCoveringController = "Window covering controller"
	DeviceTypeWindowCoveringDevice     = "Window covering device"
)

type FullState struct {
	Lights  map[string]Light       `json:"lights"`
	Sensors map[string]Sensor      `json:"sensors"`
	Groups  map[string]interface{} `json:"groups"`
}

type ConbeeEvent struct {
	MsgType      string                 `json:"t"`
	EventType    string                 `json:"e"`
	ResourceType string                 `json:"r"`
	Id           string                 `json:"id"`
	State        map[string]interface{} `json:"state"`
	Light        Light                  `json:"light"`
	Sensor       Sensor                 `json:"sensor"`
}

type Light struct {
	Etag             string `json:"etag"`
	Hascolor         bool   `json:"hascolor"`
	Manufacturername string `json:"manufacturername"`
	Modelid          string `json:"modelid"`
	Name             string `json:"name"`
	Pointsymbol      struct {
	} `json:"pointsymbol"`
	State struct {
		Alert     string    `json:"alert"`
		Bri       int       `json:"bri"`
		Colormode string    `json:"colormode"`
		Ct        int       `json:"ct"`
		Effect    string    `json:"effect"`
		Hue       int       `json:"hue"`
		On        bool      `json:"on"`
		Reachable bool      `json:"reachable"`
		Sat       int       `json:"sat"`
		Xy        []float64 `json:"xy"`
		Open 	  *bool `json:"open,omitempty"`
		Stop      *bool `json:"stop,omitempty"`
		Lift      *int  `json:"lift,omitempty"`
	} `json:"state"`
	Swversion string `json:"swversion"`
	Type      string `json:"type"`
	Uniqueid  string `json:"uniqueid"`
}

type Sensor struct {
	Config struct {
		On        bool  `json:"on"`
		Reachable *bool `json:"reachable"`
		Battery   *int  `json:"battery"`
	} `json:"config"`
	Ep               int    `json:"ep"`
	Etag             string `json:"etag"`
	Manufacturername string `json:"manufacturername"`
	Modelid          string `json:"modelid"`
	Name             string `json:"name"`
	State            struct {
		Lastupdated string `json:"lastupdated"`
	} `json:"state"`
	Swversion string `json:"swversion"`
	Type      string `json:"type"`
	Uniqueid  string `json:"uniqueid"`
}

type ConnbeeLightRequest struct {
	On  bool `json:"on"`
	Bri int  `json:"bri,omitempty"`
}

type WindowCoveringRequest struct {
	Open *bool `json:"open,omitempty"`
	Stop *bool `json:"stop,omitempty"`
	Lift *int  `json:"lift,omitempty"`
}
