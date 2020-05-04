package mqtt

//Client implements the mqtt client
type Client struct{}

//packet
type packet struct {
	fixedHeader    []byte
	variableHeader []byte
	payload        []byte
}
type client interface {
	Connect()
	Publish()
	AcknowledgedPublish()
	ReceivedPublish()
	ReleasePublish()
	CompletedPublish()
	Subscribe()
	Unsubscribe()
	Ping()
	Disconnect()
}

//Connect
func (c *Client) Connect() {
	pkt := packet{}
	pkt.fixedHeader[0] = 0x10
	//Protocol Name
	pkt.variableHeader[0] = byte(0)
	pkt.variableHeader[1] = byte(4)
	pkt.variableHeader[2] = byte('M')
	pkt.variableHeader[3] = byte('Q')
	pkt.variableHeader[4] = byte('T')
	pkt.variableHeader[5] = byte('T')
	//Protocol Level
	pkt.variableHeader[6] = byte(4)
}
