package model

import "github.com/futurehomeno/fimpgo/discovery"

func GetDiscoveryResource() discovery.Resource {
	//adInterfaces := []fimptype.Interface{{
	//	Type:      "in",
	//	MsgType:   "cmd.network.get_all_nodes",
	//	ValueType: "null",
	//	Version:   "1",
	//}, {
	//	Type:      "in",
	//	MsgType:   "cmd.thing.get_inclusion_report",
	//	ValueType: "string",
	//	Version:   "1",
	//}, {
	//	Type:      "in",
	//	MsgType:   "cmd.thing.inclusion",
	//	ValueType: "string",
	//	Version:   "1",
	//}, {
	//	Type:      "in",
	//	MsgType:   "cmd.thing.delete",
	//	ValueType: "string",
	//	Version:   "1",
	//}, {
	//	Type:      "in",
	//	MsgType:   "cmd.auth.login",
	//	ValueType: "str_map", // username/password
	//	Version:   "1",
	//}, {
	//	Type:      "out",
	//	MsgType:   "evt.auth.login_report",
	//	ValueType: "string", // success , failed
	//	Version:   "1",
	//}, {
	//	Type:      "out",
	//	MsgType:   "evt.thing.inclusion_report",
	//	ValueType: "object",
	//	Version:   "1",
	//}, {
	//	Type:      "out",
	//	MsgType:   "evt.thing.exclusion_report",
	//	ValueType: "object",
	//	Version:   "1",
	//}, {
	//	Type:      "out",
	//	MsgType:   "evt.network.all_nodes_report",
	//	ValueType: "object",
	//	Version:   "1",
	//}, {
	//	Type:      "in",
	//	MsgType:   "cmd.log.set_level",
	//	ValueType: "string",
	//	Version:   "1",
	//}}
	//
	//adService := fimptype.Service{
	//	Name:             "conbee",
	//	Alias:            "Network managment",
	//	Address:          "/rt:ad/rn:conbee/ad:1",
	//	Enabled:          true,
	//	Groups:           []string{"ch_0"},
	//	Tags:             nil,
	//	PropSetReference: "",
	//	Interfaces:       adInterfaces,
	//}
	return discovery.Resource{
		ResourceName:           "conbee",
		ResourceType:           discovery.ResourceTypeAd,
		Author:                 "aleksandrs.livincovs@gmail.com",
		IsInstanceConfigurable: false,
		InstanceId:             "1",
		Version:                "1",
		AdapterInfo: discovery.AdapterInfo{
			Technology:            "conbee",
			FwVersion:             "all",
			NetworkManagementType: "inclusion_exclusion",
		},
	}

}
