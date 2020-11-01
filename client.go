package mqtt

import (
	"fmt"
	"io"
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
//Connect configure and send the connect packet to the server
func (c *Client) Connect() {
	pkt := packet{}
	pkt.configureConnectPackets(c.options)
	sendBytes:=pkt.FormulateMQTTOutputData()
	_,_=fmt.Fprintf(c.conn,"%s",string(sendBytes))
	buf:=make([]byte,0,40960)
	tmp:=make([]byte,1)
	for {
		status,err:=c.conn.Read(tmp)
		if err!=nil{
		  if err==io.EOF{
		  	fmt.Println("EOF")
		  	break
		  }
		 fmt.Println(err.Error())
		  }
		  buf = append(buf,tmp[:status]...)
		  fmt.Print(tmp[0])
	}
	defer c.conn.Close()
}
