package mqtt

import (
	"encoding/binary"
	"log"
	"fmt"
	"net"
	"sync"
	"time"
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
	done chan int
	commLock sync.Mutex
	topic string
	qos QOS
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
	c.commLock.Lock()
	_,_=c.conn.Write(sendBytes)
	c.commLock.Unlock()
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
		case ConnectionRefusedIdentifierRejected:
			return &mqttErr{ConnectionRefusedIdentifierRejected,"Identifier rejected"}
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
	//defer c.conn.Close()
	return nil
}

func(c *Client) AwaitMessages(){
	b:=make([]byte,1)
	var i int
	for{
		c.conn.SetReadDeadline(time.Now().Add(1*time.Millisecond))
		c.commLock.Lock()
		i,_=c.conn.Read(b)
		c.commLock.Unlock()
		if i>0{
			fmt.Print(string(b))
		}
	}
}

func (c *Client) SendPing(){
	pkt:=packet{}
	pkt.configurePingRequest()
	pingBytes:=pkt.FormulateMQTTOutputData()
	d:= time.Duration(c.options.KeepAlive)
	for {
		select {
		case d:=<-c.done:
			fmt.Println("Done",d)
			c.conn.Close()
		default:
			c.commLock.Lock()
			_,err:=c.conn.Write(pingBytes)
			if err!=nil{
				fmt.Println(err.Error())
				c.done<-1
			}
			resp:=make([]byte,2)
			c.conn.SetReadDeadline(time.Now().Add(3*time.Second))
			_,err=c.GetSrvPacket(resp)
			if err!=nil{
				fmt.Print(err.Error())
			}
			c.commLock.Unlock()
			time.Sleep(d*time.Second)
		}

	}
}
func (c *Client) Subscribe(topic string, qos QOS) error{
	pkt:=packet{}
	pkt.configureSubscribePackets(topic,qos)
	c.topic = topic
	c.qos = qos
	sendBytes:=pkt.FormulateMQTTOutputData()
	c.commLock.Lock()
	c.conn.Write(sendBytes)
	c.commLock.Unlock()
	ctrl:=make([]byte,1)
	r,err:=c.GetSrvPacket(ctrl)
	if err!=nil{
		return &mqttErr{r,"Issue with getting packet"}
	}
	if int(ctrl[0])==ControlPktSubAck{
		r,err=c.GetSrvPacket(ctrl)
		remLength:=int(ctrl[0])
		data:=make([]byte,remLength)
		_,err=c.GetSrvPacket(data)
		if err!=nil{
			return &mqttErr{r,"Issue with getting packet"}
		}
		pktId:=binary.BigEndian.Uint16(data[0:])
		if pktId !=PacketIdentifier{
			return &mqttErr{int(pktId),"Packet identifier mismatch"}
		}
		if byte(qos)!=data[2]{
			return &mqttErr{int(data[2]),"QOS mismatch"}
		}
	}else{
		return &mqttErr{int(ctrl[0]),"Failed to subscribe"}
	}
	go c.SendPing()
	go c.AwaitMessages()
	fmt.Println("Sent Ping")
	return nil
}
func (c *Client) Publish(message string) error{
	pkt:=packet{}
	pkt.configurePublish(c.topic,message,false,c.qos)
	sendBytes:=pkt.FormulateMQTTOutputData()
	c.commLock.Lock()
	c.conn.Write(sendBytes)

	c.commLock.Unlock()
	return nil
}