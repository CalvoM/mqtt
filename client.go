package mqtt

import (
	"fmt"
	"log"
	"net"
)

const (
	MqttProtocolLvl = 4
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
func (c *Client) Connect() error{
	pkt := packet{}
	pkt.configureConnectPackets(c.options)
	sendBytes:=pkt.FormulateMQTTOutputData()
	_,_=fmt.Fprintf(c.conn,"%s",string(sendBytes))
	ctrl:=make([]byte,1)
	r,err:=c.GetSrvPacket(ctrl)
	if err!=nil{
		return &mqttErr{r,"Issue with getting packet"}
	}
	if int(ctrl[0])==ControlPktConnAck{
		data:=make([]byte,3)
		c.GetSrvPacket(data)
		switch data[2] {
		case ConnectionAccepted:
			return nil
		case ConnectionRefusedProtocolVersion:
			return &mqttErr{ConnectionRefusedProtocolVersion,"Protocol Version Not supported"}
		case ConnectionRefusedServerUnavailable:
			return &mqttErr{ConnectionRefusedServerUnavailable,"Server Unavailable"}
		case ConnectionRefusedUsernamePassword:
			return &mqttErr{ConnectionRefusedUsernamePassword,"Wrong Username or Password"}
		case ConnectionRefusedNotAuthorized:
			return &mqttErr{ConnectionRefusedNotAuthorized,"Not Authorised"}
		}
	}else{
		return &mqttErr{-1,"Connack Packet Not received"}
	}
	defer c.conn.Close()
	return nil
}
