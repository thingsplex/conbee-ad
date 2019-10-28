package model

import (
	log "github.com/sirupsen/logrus"
	"sync"
)


//
// Events : STATING -> CONFIGURING -> CONFIGURED -> RUNNING
// States : CONFIGURING -> RUNNING  / CONFIGURING -> NOT_CONFIGURED /

const (
	SystemEventTypeEvent = "EVENT"
	SystemEventTypeState = "STATE"

	StateConfiguring         = "CONFIGURING" // Brokers configured
	StateConfigured          = "CONFIGURED" // Brokers configured
	StateNotConfigured       = "NOT_CONFIGURED" // Brokers configured
	StateRunning             = "RUNNING"
	StateStartupError        = "STARTUP_ERROR"
	StateTerminate           = "TERMINATE"

	//EventStarting            = "STARTING"
	EventConfiguring         = "CONFIGURING"          // All configurations loaded and brokers configured
	EventConfigError         = "CONF_ERROR"           // All configurations loaded and brokers configured
	EventConfigured          = "CONFIGURED"          // All configurations loaded and brokers configured
	EventRunning             = "RUNNING"

)

type State string

type SystemEvent struct {
	Type   string
	Name   string
	State  State
	Info   string
	Params map[string]string
}

type SystemEventChannel chan SystemEvent

type Lifecycle struct {
	busMux                 sync.Mutex
	systemEventBus         map[string]SystemEventChannel
	currentState           State
	previousState          State
	cloudConnectivityState State
	receiveChTimeout       int
	backendVersion         int
	isPaired               bool
}

func NewAppLifecycle() *Lifecycle {

	return &Lifecycle{systemEventBus: make(map[string]SystemEventChannel)}
}

func (al *Lifecycle) CurrentState() State {
	return al.currentState
}

func (al *Lifecycle) SetCurrentState(currentState State, params map[string]string) {
	al.busMux.Lock()
	al.previousState = al.currentState
	al.currentState = currentState
	log.Debug("<sysEvt> New system state = ", currentState)
	for i := range al.systemEventBus {
		select {
		case al.systemEventBus[i] <- SystemEvent{Type: SystemEventTypeState, State: currentState, Info: "sys", Params: params}:
		default:
			log.Warnf("<sysEvt> State listener %s is busy , event dropped", i)
		}

	}
	al.busMux.Unlock()
}

func (al *Lifecycle) PublishEvent(name,src string, params map[string]string) {
	event := SystemEvent{Name:name}
	al.Publish(event,src,params)
}

func (al *Lifecycle) Publish(event SystemEvent, src string, params map[string]string) {
	al.busMux.Lock()
	event.Type = SystemEventTypeEvent
	event.State = al.CurrentState()
	event.Params = params
	for i := range al.systemEventBus {
		select {
		case al.systemEventBus[i] <- event:
		default:
			log.Warnf("<sysEvt> Event listener %s is busy , event dropped", i)
		}

	}
	al.busMux.Unlock()
	al.processEvent(event)
}

func (al *Lifecycle) Subscribe(subId string, bufSize int) SystemEventChannel {
	msgChan := make(SystemEventChannel, bufSize)
	al.busMux.Lock()
	al.systemEventBus[subId] = msgChan
	al.busMux.Unlock()
	return msgChan
}

func (al *Lifecycle) Unsubscribe(subId string) {
	al.busMux.Lock()
	delete(al.systemEventBus, subId)
	al.busMux.Unlock()
}

func (al *Lifecycle) processEvent(event SystemEvent) {
	switch event.Name {

	case EventConfiguring:
		al.SetCurrentState(StateConfiguring, nil)

	case EventConfigured:
		al.SetCurrentState(StateConfigured, nil)
		al.SetCurrentState(StateRunning, nil)

	case EventConfigError:
		al.SetCurrentState(StateNotConfigured, nil)
	}

}

// WaitForState blocks until target state is reached
func (al *Lifecycle) WaitForState(subId string, targetState State) {
	log.Debugf("<sysEvt> Waiting from event = %s , current state = %s", targetState, al.CurrentState())
	if al.CurrentState() == targetState {
		return
	}
	ch := al.Subscribe(subId, 5)
	for evt := range ch {
		if evt.Type == SystemEventTypeState && evt.State == targetState {
			al.Unsubscribe(subId)
			return
		}
	}
}
