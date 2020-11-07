package mqtt

import "encoding/binary"

//packet
type packet struct {
	fixedHeader    []byte
	variableHeader []byte
	payload        []byte
}
//ConnectOptions These correspond to the connect flags
type ConnectOptions struct {
	Username     string
	Password     string
	WillRetain   bool
	WillQOS      QOS
	WillFlag     bool
	CleanSession bool
	KeepAlive    uint16
	WillTopic	 string
	WillMessage string
	ClientId string
}
//FormulateMQTTOutputData
func(pkt *packet) FormulateMQTTOutputData()[]byte{
	var s []byte
	s = append(s,pkt.fixedHeader...)
	s = append(s,pkt.variableHeader...)
	s = append(s,pkt.payload...)
	return s
}

func (pkt *packet) configureConnectPackets(options *ConnectOptions) {
	pkt.fixedHeader = append(pkt.fixedHeader,ControlPktConnect)
	//Protocol Name
	pkt.variableHeader = append(pkt.variableHeader,0x0)
	pkt.variableHeader = append(pkt.variableHeader,byte(len("MQTT")))
	pkt.variableHeader=append(pkt.variableHeader,[]byte("MQTT")...)
	//Protocol Level
	pkt.variableHeader= append(pkt.variableHeader,byte(MqttProtocolLvl))
	//Connect Flags
	connectFlag := byte(0)
	if options.CleanSession {
		connectFlag |= FlagCleanSessionEnabled
	}
	if options.WillFlag {
		connectFlag |= FlagWillFlagEnabled
		if options.WillQOS == 0 {
			connectFlag |= 0x00
		}
		if options.WillQOS == QOS1 {
			connectFlag |= FlagWillQOS1
		}
		if options.WillQOS == QOS2 {
			connectFlag |= FlagWillQOS2
		}
		if options.WillRetain {
			connectFlag |= FlagWillRetain
		}
	}
	if options.Username != "" {
		connectFlag |= FlagUsername
		if options.Password != "" {
			connectFlag |= FlagPassword
		}
	}
	pkt.variableHeader=append(pkt.variableHeader,connectFlag)
	//KeepAlive bytes
	var keepAlive uint16 = options.KeepAlive
	k:=make([]byte,2)
	binary.BigEndian.PutUint16(k,keepAlive)
	pkt.variableHeader = append(pkt.variableHeader,k...)
	//configure the payload
	pkt.SetConnectPayload(options)
	length:=len(pkt.variableHeader)+len(pkt.payload)
	remLength:=encodeRemainingLength(uint64(length))
	pkt.fixedHeader = append(pkt.fixedHeader,remLength...)
}

func(pkt *packet) SetConnectPayload(options *ConnectOptions){
	var data uint16 = uint16(len(options.ClientId))
	b :=make([]byte,2)
	binary.BigEndian.PutUint16(b,data)
	pkt.payload = append(pkt.payload,b...)
	pkt.payload = append(pkt.payload,[]byte(options.ClientId)...)
	if options.WillFlag{
		data = uint16(len(options.WillTopic))
		binary.BigEndian.PutUint16(b,data)
		pkt.payload = append(pkt.payload,b...)
		pkt.payload = append(pkt.payload,[]byte(options.WillTopic)...)
		data = uint16(len(options.WillMessage))
		binary.BigEndian.PutUint16(b,data)
		pkt.payload = append(pkt.payload,b...)
		pkt.payload = append(pkt.payload,[]byte(options.WillMessage)...)
	}
	if len(options.Username)>0{
		data = uint16(len(options.Username))
		binary.BigEndian.PutUint16(b,data)
		pkt.payload = append(pkt.payload,b...)
		pkt.payload = append(pkt.payload,[]byte(options.Username)...)
		if len(options.Password)>0{
			data = uint16(len(options.Password))
			binary.BigEndian.PutUint16(b,data)
			pkt.payload = append(pkt.payload,b...)
			pkt.payload = append(pkt.payload,[]byte(options.Password)...)
		}
	}

}
func (pkt *packet) configureSubscribePackets(topic string, qos QOS){
	pkt.fixedHeader = append(pkt.fixedHeader,ControlPktSubscribe|0x02)
	var packetId uint16 = PacketIdentifier
	k:=make([]byte,2)
	binary.BigEndian.PutUint16(k,packetId)
	pkt.variableHeader = append(pkt.variableHeader,k...)
	n:=make([]byte,2)
	var topicLen uint16 = uint16(len(topic))
	binary.BigEndian.PutUint16(n,topicLen)
	pkt.payload = append(pkt.payload,n...)
	pkt.payload = append(pkt.payload,[]byte(topic)...)
	pkt.payload = append(pkt.payload,byte(qos))
	length:=len(pkt.variableHeader)+len(pkt.payload)
	remLength:=encodeRemainingLength(uint64(length))
	pkt.fixedHeader = append(pkt.fixedHeader,remLength...)
}