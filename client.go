package mqtt

import (
	"fmt"
	"log"
	"net"
)

const (
	mqttProtocolLvl = 4
)

//Client implements the mqtt client
type Client struct{
	Host string
	Port string
	options *ConnectOptions
	conn net.Conn
}


type client interface {
	Connect(options *ConnectOptions)
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

//Init initialize the client options and dial connection up
func (c *Client) Init(options *ConnectOptions){
	var err error
	c.options = options
	addr:=net.JoinHostPort(c.Host,c.Port)
	c.conn,err=net.Dial("tcp",addr)
	if err!=nil{
		log.Fatal(err)
	}
}
func (c *Client) GetSrvPacket(buf []byte)(int,error){
	return c.conn.Read(buf)
}
//Connect configure and send the connect packet to the server
func (c *Client) Connect() {
	pkt := packet{}
	pkt.configureConnectPackets(c.options)
	sendBytes:=pkt.FormulateMQTTOutputData()
	_,_=fmt.Fprintf(c.conn,"%s",string(sendBytes))
	ctrl:=make([]byte,1)
	c.GetSrvPacket(ctrl)
	if int(ctrl[0])==ControlPktConnAck{
		data:=make([]byte,3)
		c.GetSrvPacket(data)
		fmt.Println("Remaining Length:",data[0])
		fmt.Println("Session Present:",data[1])
		fmt.Println("Connect Return Code:",data[2])
	}
	defer c.conn.Close()
}
