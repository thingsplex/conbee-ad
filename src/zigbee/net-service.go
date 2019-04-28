package zigbee

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/conbee-ad/conbee"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	"github.com/pkg/errors"
)

type ListReportRecord struct {
	Address     string `json:"address"`
	Alias       string `json:"alias"`
	PowerSource string `json:"power_source"`
	Hash        string `json:"hash"`
}

type NetworkService struct {
	mqt          *fimpgo.MqttTransport
	conbeeClient *conbee.Client
}

func NewNetworkService(mqt *fimpgo.MqttTransport, conbeeClient *conbee.Client) *NetworkService {
	return &NetworkService{mqt: mqt, conbeeClient: conbeeClient}
}

func (ns *NetworkService) OpenNetwork(open bool) {

}

func (ns *NetworkService) DeleteThing(deviceId string) {

}

func (ns *NetworkService) SendInclusionReport(deviceType, deviceId string) error {
	if deviceId == "" {
		return errors.New("empty device id")
	}
	var productId, manufacturer, powerSource, swVersion, serialNr string
	var tryNextDev bool
	services := []fimptype.Service{}

	outLvlSwitchInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.binary.set",
		ValueType: "bool",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.set",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.start",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.stop",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.lvl.report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.binary.report",
		ValueType: "bool",
		Version:   "1",
	}}


	outLvlSwitchService := fimptype.Service{
		Name:    "out_lvl_switch",
		Alias:   "Light control",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:out_lvl_switch/ad:"+deviceId+"_0",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"max_lvl": 255,
			"min_lvl": 0,
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       outLvlSwitchInterfaces,
	}



	if deviceType == "light" || deviceType == "" {
		lightDeviceDescriptor := conbee.Light{}
		resp , err := ns.conbeeClient.SendConbeeRequest("GET", "lights/"+deviceId, nil, &lightDeviceDescriptor)
		if resp.StatusCode == 404 {
			tryNextDev = true
		} else 	if err != nil {
			log.Error("Can't get device descriptor . Err :", err)
			return err
		}

		productId = lightDeviceDescriptor.Modelid
		manufacturer = lightDeviceDescriptor.Manufacturername
		swVersion = lightDeviceDescriptor.Swversion
		serialNr = lightDeviceDescriptor.Uniqueid
		services = append(services,outLvlSwitchService)
	}
	if deviceType == "sensor" || tryNextDev {
		sensorDeviceDescriptor := conbee.Sensor{}
		_ , err := ns.conbeeClient.SendConbeeRequest("GET", "sensors/"+deviceId, nil, &sensorDeviceDescriptor)
		if err != nil {
			log.Error("Can't get device descriptor . Err :", err)
			return err
		}
		productId = sensorDeviceDescriptor.Modelid
		manufacturer = sensorDeviceDescriptor.Manufacturername
		swVersion = sensorDeviceDescriptor.Swversion
		serialNr = sensorDeviceDescriptor.Uniqueid
	}


	powerSource = ""
	inclReport := fimptype.ThingInclusionReport{
		IntegrationId:     "",
		Address:           deviceId,
		Type:              "",
		ProductHash:       manufacturer + "_" + productId,
		Alias:             productId,
		CommTechnology:    "zigbee",
		ProductId:         productId,
		ProductName:       productId,
		ManufacturerId:    manufacturer,
		DeviceId:          serialNr,
		HwVersion:         "1",
		SwVersion:         swVersion,
		PowerSource:       powerSource,
		WakeUpInterval:    "-1",
		Security:          "",
		Tags:              nil,
		Groups:            []string{"ch_0"},
		PropSets:          nil,
		TechSpecificProps: nil,
		Services:          services,
	}

	msg := fimpgo.NewMessage("evt.thing.inclusion_report", "zigbee", fimpgo.VTypeObject, inclReport, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "zigbee", ResourceAddress: "1"}
	ns.mqt.Publish(&adr, msg)
	return nil

}

func (ns *NetworkService) SendListOfDevices() error {
	lights := map[string]conbee.Light{}
	sensors := map[string]conbee.Sensor{}
	_, err := ns.conbeeClient.SendConbeeRequest("GET", "lights", nil, &lights)
	if err != nil {
		log.Error("Can't get device descriptor . Err :", err)
		return err
	}
	//err = json.NewDecoder(resp.Body).Decode(&lights)
	//if err != nil {
	//	log.Error("Can't decode device descriptor . Err :", err)
	//	return err
	//}

	_, err = ns.conbeeClient.SendConbeeRequest("GET", "sensors", nil, &sensors)
	if err != nil {
		log.Error("Can't get device descriptor . Err :", err)
		return err
	}
	//err = json.NewDecoder(resp.Body).Decode(&sensors)
	//if err != nil {
	//	log.Error("Can't decode device descriptor . Err :", err)
	//	return err
	//}

	report := []ListReportRecord{}

	for i := range lights {
		rec := ListReportRecord{Address:i,Alias:lights[i].Name}
		report = append(report,rec)
	}
	for i := range sensors {
		rec := ListReportRecord{Address:i,Alias:sensors[i].Name}
		report = append(report,rec)
	}
	msg := fimpgo.NewMessage("evt.network.all_nodes_report", "zigbee", fimpgo.VTypeObject, report, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "zigbee", ResourceAddress: "1"}
	ns.mqt.Publish(&adr, msg)

	return nil
}
