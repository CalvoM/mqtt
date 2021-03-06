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

	FlagCleanSessionEnabled=0x02
	FlagWillFlagEnabled=0x04
	FlagWillQOS1=0x08
	FlagWillQOS2=0x10
	FlagWillRetain=0x20
	FlagUsername=0x80
	FlagPassword=0x40

	ConnectionAccepted=0x00
	ConnectionRefusedProtocolVersion=0x01
	ConnectionRefusedIdentifierRejected=0x02
	ConnectionRefusedServerUnavailable=0x03
	ConnectionRefusedUsernamePassword=0x04
	ConnectionRefusedNotAuthorized=0x05
	PacketIdentifier=0x0010
)
type QOS int8
const (
	QOS0 QOS = iota
	QOS1
	QOS2
	QOSFail
)
type MessageHandler func(MessageData)
type MessageData struct{
	Topic string
	Payload string
}