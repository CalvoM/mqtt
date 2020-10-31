package mqtt
const(
	ControlPktConnect=0x10
	ControlPktConnAck=0x20
	ControlPktPublish=0x30
	ControlPktPubAck=0x40
	ControlPktPubRec=0x50
	ControlPktPubRel=0x60
	ControlPktPubComp=0x70
	ControlPktSubscribe=0x80
	ControlPktSubAck=0x90
	ControlPktUnsubscribe=0xA0
	ControlPktUnsubAck=0xB0
	ControlPktPingReq=0xC0
	ControlPktPingResp=0xD0
	ControlPktDisconnect=0xE0
	ControlPktAuth=0xF0
)