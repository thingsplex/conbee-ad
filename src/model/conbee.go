package model

type ConnbeeLightRequest struct {
	On bool `json:"on"`
	Bri int `json:"bri,omitempty"`

}