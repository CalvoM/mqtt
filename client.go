package mqtt

import (
	"fmt"
	"io"
	"log"
	"math"
	"net"
)

const (
	mqttProtocolLvl = 4
)

//Client implements the mqtt client
type Client struct{
	Username string
	Password string
	WillQOS int
	ClientId string
	WillTopic string
	Host string
	Port string
}

//packet
type packet struct {
	fixedHeader    []byte
	variableHeader []byte
	payload        []byte
}

func decodeRemainingLength(encodedBytes []byte) (uint64,error){
	var multiplier uint64=1
	var value uint64 =0
	i:=0
	encodedByte:=byte(1)
	for (encodedByte&0x80)!=0{
		encodedByte=encodedBytes[i]
		value+=uint64(uint64(encodedByte&0x7F)*multiplier)
		multiplier*=0x80
		if float64(multiplier)>math.Pow(128,3){
			return 0,io.ErrNoProgress
		}
		i+=1
	}
	return value,nil
}

func encodeRemainingLength(n uint64) []byte{
	var encodedByte byte
	encodedBytes:=make([]byte,0)
	for n>0{
			encodedByte = byte(n%0x80)
			n = uint64(n/0x80)
			if n>0{
				encodedByte = encodedByte | 0x80
			}
			encodedBytes=append(encodedBytes,encodedByte)
	}
	return encodedBytes
}

func (pkt *packet) configureConnectHeaders(options *ConnectOptions) {
	var payload string
	payload+=options.ClientId
	pkt.fixedHeader[0] = 0x10
	//Protocol Name
	pkt.variableHeader[0] = byte(0)
	pkt.variableHeader[1] = byte(len("MQTT"))
	pkt.variableHeader[2] = byte('M')
	pkt.variableHeader[3] = byte('Q')
	pkt.variableHeader[4] = byte('T')
	pkt.variableHeader[5] = byte('T')
	//Protocol Level
	pkt.variableHeader[6] = byte(mqttProtocolLvl)
	//Connect Flags
	connectFlag := byte(0)
	if options.CleanSession {
		connectFlag |= 0x02
	}
	if options.WillFlag {
		payload+=options.WillTopic
		payload+=options.WillMessage
		connectFlag |= 0x04
		if options.WillQOS == 0 {
			connectFlag |= 0x00
		}
		if options.WillQOS == 1 {
			connectFlag |= 0x08
		}
		if options.WillQOS == 2 {
			connectFlag |= 0x10
		}
		if options.WillRetain {
			connectFlag |= 0x20
		}
	}
	if options.Username != "" {
		payload+=options.Username
		connectFlag |= 0x80
		if options.Password != "" {
			payload+=options.Password
			connectFlag |= 0x40
		}
	}
	pkt.variableHeader[7]=connectFlag
	//KeepAlive bytes
	pkt.variableHeader[8] = 0x00
	pkt.variableHeader[9] = 0x00
	pkt.payload=[]byte(payload)
	length:=len(pkt.variableHeader)+len(pkt.payload)
	remLength:=encodeRemainingLength(uint64(length))
	pkt.fixedHeader = append(pkt.fixedHeader,remLength...)
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

//ConnectOptions These correspond to the connect flags
type ConnectOptions struct {
	Username     string
	Password     string
	WillRetain   bool
	WillQOS      int
	WillFlag     bool
	CleanSession bool
	KeepAlive    int
	WillTopic	 string
	WillMessage string
	ClientId string
}

//Connect configure and send the connect packet to the server
func (c *Client) Connect(options *ConnectOptions) {
	pkt := packet{}
	c.ClientId = options.ClientId
	c.Password = options.Password
	c.Username = options.Username
	c.WillQOS = options.WillQOS
	c.WillTopic = options.WillTopic
	pkt.configureConnectHeaders(options)
	addr:=net.JoinHostPort(c.Host,c.Port)
	conn,err:= net.Dial("tcp",addr)
	if err!=nil{
		log.Fatal(err)
	}
	defer conn.Close()
	var sendBytes []byte
	sendBytes = append(sendBytes,pkt.fixedHeader...)
	sendBytes = append(sendBytes,pkt.variableHeader...)
	sendBytes = append(sendBytes,pkt.payload...)
	_,_=fmt.Fprint(conn,sendBytes)
	localIP,_,_:=net.SplitHostPort(conn.LocalAddr().String())
	remoteIP,_,_:=net.SplitHostPort(conn.RemoteAddr().String())
	fmt.Printf("Local Address: %s\r\n",localIP)
	fmt.Printf("Remote Address:%s\r\n",remoteIP)
	buf:=make([]byte,0,40960)
	tmp:=make([]byte,1)
	for {
		status,err:=conn.Read(tmp)
		if err!=nil{
		  if err==io.EOF{
		  	break
		  }
		 fmt.Println(err.Error())
		  }
		  buf = append(buf,tmp[:status]...)
		  fmt.Print(tmp)
	}
}
