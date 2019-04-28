package zigbee

import (
	log "github.com/sirupsen/logrus"
	"github.com/alivinco/conbee-ad/conbee"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	"github.com/pkg/errors"
	"strings"
)

type ListReportRecord struct {
	Address     string `json:"address"`
	Alias       string `json:"alias"`
	PowerSource string `json:"power_source"`
	Hash        string `json:"hash"`
}

type OpResponse struct {

}

type NetworkService struct {
	mqt          *fimpgo.MqttTransport
	conbeeClient *conbee.Client
}

func NewNetworkService(mqt *fimpgo.MqttTransport, conbeeClient *conbee.Client) *NetworkService {
	return &NetworkService{mqt: mqt, conbeeClient: conbeeClient}
}

func (ns *NetworkService) OpenNetwork(open bool) error {
	var val int
	if open {
		val = 120
	}
	req := map[string]int{"permitjoin":val}
	resp, err := ns.conbeeClient.SendConbeeRequest("PUT", "config", req, nil)
	if err != nil {
		log.Error("Can open or close network . Err :", err)
		return err
	}
	if resp.StatusCode == 200 {
		if open {
			log.Info("Network is open")
		}else {
			log.Info("Network is closed")

		}
	}else {
		log.Error("Network managment failure .")
		return errors.New("non success status code")
	}
	return nil
}

func (ns *NetworkService) DeleteThing(deviceId string) error {
	var path string
	if strings.Contains(deviceId,"s") {
		path = "sensors/"+strings.Replace(deviceId,"s","",1)
	}else {
		path = "lights/"+strings.Replace(deviceId,"l","",1)
	}
	req := map[string]bool{"reset":true}
	resp, err := ns.conbeeClient.SendConbeeRequest("DELETE", path, req, nil)
	if err != nil {
		log.Error("Can't delete device from network . Err :", err)
		return err
	}
	if resp.StatusCode == 200 {
			log.Info("Device deleted")
	}else {
		log.Error("Network managment failure .")
		return errors.New("non success status code")
	}
	exclReport := map[string]string{"address":deviceId}
	msg := fimpgo.NewMessage("evt.thing.exclusion_report", "zigbee", fimpgo.VTypeObject, exclReport, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "zigbee", ResourceAddress: "1"}
	ns.mqt.Publish(&adr, msg)

	return nil

}

func (ns *NetworkService) SendInclusionReport(deviceType, deviceId string) error {
	if deviceId == "" {
		return errors.New("empty device id")
	}
	var productId, manufacturer, powerSource, swVersion, serialNr string
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

	batteryInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.lvl.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.lvl.report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.alarm.report",
		ValueType: "str_map",
		Version:   "1",
	}}

	sensorInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.sensor.get_report",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.sensor.report",
		ValueType: "float",
		Version:   "1",
	}}

	presenceSensorInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.presence.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.presence.report",
		ValueType: "bool",
		Version:   "1",
	}}

	contactSensorInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.open.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.open.report",
		ValueType: "bool",
		Version:   "1",
	}}

	sceneInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.scene.set",
		ValueType: "string",
		Version:   "1",
	},{
		Type:      "in",
		MsgType:   "cmd.scene.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.scene.report",
		ValueType: "string",
		Version:   "1",
	}}



	outLvlSwitchService := fimptype.Service{
		Name:    "out_lvl_switch",
		Alias:   "Light control",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:out_lvl_switch/ad:",
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

	tempSensorService := fimptype.Service{
		Name:    "sensor_temp",
		Alias:   "Temperature sensor",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:sensor_temp/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units": []string{"C"},
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       sensorInterfaces,
	}

	humidSensorService := fimptype.Service{
		Name:    "sensor_humid",
		Alias:   "Humidity sensor",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:sensor_humid/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units": []string{"%"},
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       sensorInterfaces,
	}

	batteryService := fimptype.Service{
		Name:    "battery",
		Alias:   "battery",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:battery/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       batteryInterfaces,
	}

	sceneService := fimptype.Service{
		Name:    "scene_ctrl",
		Alias:   "Scene",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:scene_ctrl/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       sceneInterfaces,
	}

	contactService := fimptype.Service{
		Name:    "sensor_contact",
		Alias:   "Door/window contact",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:sensor_contact/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       contactSensorInterfaces,
	}

	presenceService := fimptype.Service{
		Name:    "sensor_presence",
		Alias:   "Door/window contact",
		Address: "/rt:dev/rn:zigbee/ad:1/sv:sensor_presence/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       presenceSensorInterfaces,
	}

	if deviceType == "lights" {
		lightDeviceDescriptor := conbee.Light{}
		resp , err := ns.conbeeClient.SendConbeeRequest("GET", "lights/"+deviceId, nil, &lightDeviceDescriptor)
		if resp.StatusCode == 404 {
			log.Error("Device not found . Err :", err)
			return err
		} else 	if err != nil {
			log.Error("Can't get device descriptor . Err :", err)
			return err
		}else {
			// All good
			productId = lightDeviceDescriptor.Modelid
			manufacturer = lightDeviceDescriptor.Manufacturername
			swVersion = lightDeviceDescriptor.Swversion
			serialNr = lightDeviceDescriptor.Uniqueid
			serviceAddres := "l"+deviceId+"_0"
			outLvlSwitchService.Address = outLvlSwitchService.Address + serviceAddres
			services = append(services,outLvlSwitchService)
		}
		powerSource = "ac"
		deviceId = "l"+deviceId

	}
	if deviceType == "sensors" {
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

		serviceAddres := "s"+deviceId+"_0"
		batteryService.Address = batteryService.Address + serviceAddres
		services = append(services,batteryService)

		switch sensorDeviceDescriptor.Type {
		case "ZHATemperature":
			tempSensorService.Address = tempSensorService.Address + serviceAddres
			services = append(services,tempSensorService)
		case "ZHAHumidity":
			humidSensorService.Address = humidSensorService.Address + serviceAddres
			services = append(services,humidSensorService)
		case "ZHASwitch":
			sceneService.Address = sceneService.Address + serviceAddres
			services = append(services,sceneService)
		case "ZHAOpenClose":
			contactService.Address = contactService.Address + serviceAddres
			services = append(services,contactService)
		case "ZHAPresence":
			presenceService.Address = presenceService.Address + serviceAddres
			services = append(services,presenceService)

		}
		powerSource = "battery"
		deviceId = "s"+deviceId
	}

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

	_, err = ns.conbeeClient.SendConbeeRequest("GET", "sensors", nil, &sensors)
	if err != nil {
		log.Error("Can't get device descriptor . Err :", err)
		return err
	}

	report := []ListReportRecord{}

	for i := range lights {
		rec := ListReportRecord{Address:"l"+i,Alias:lights[i].Manufacturername+" "+lights[i].Modelid}
		report = append(report,rec)
	}
	for i := range sensors {
		rec := ListReportRecord{Address:"s"+i,Alias:sensors[i].Name+" "+sensors[i].Modelid}
		report = append(report,rec)
	}
	msg := fimpgo.NewMessage("evt.network.all_nodes_report", "zigbee", fimpgo.VTypeObject, report, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "zigbee", ResourceAddress: "1"}
	ns.mqt.Publish(&adr, msg)

	return nil
}
