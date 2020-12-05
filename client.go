package mqtt

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	MqttProtocolLvl = 4
)

//Client implements the mqtt client
type Client struct {
	Host       string
	Port       string
	options    *ConnectOptions
	conn       net.Conn
	done       chan int
	readPong   chan bool
	getSignals chan os.Signal
	commLock   sync.Mutex
	topic      string
	qos        QOS
	OnMessage  MessageHandler
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
func (c *Client) Init(options *ConnectOptions) {
	var err error
	c.options = options
	c.done = make(chan int)
	c.readPong = make(chan bool)
	c.getSignals = make(chan os.Signal)
	signal.Notify(c.getSignals, syscall.SIGINT)
	addr := net.JoinHostPort(c.Host, c.Port)
	c.conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

}
func (c *Client) GetSrvPacket(buf []byte) (int, error) {
	c.commLock.Lock()
	i, err := c.conn.Read(buf)
	c.commLock.Unlock()
	return i, err
}

//Connect configure and send the connect packet to the server
func (c *Client) Connect() error {
	pkt := packet{}
	pkt.configureConnectPackets(c.options)
	sendBytes := pkt.FormulateMQTTOutputData()
	_, _ = c.conn.Write(sendBytes)
	ctrl := make([]byte, 1)
	r, err := c.GetSrvPacket(ctrl)
	if err != nil {
		return &mqttErr{r, "Issue with getting packet"}
	}
	if int(ctrl[0]) == ControlPktConnAck {
		data := make([]byte, 3)
		c.GetSrvPacket(data)
		switch data[2] {
		case ConnectionAccepted:
			return nil
		case ConnectionRefusedProtocolVersion:
			return &mqttErr{ConnectionRefusedProtocolVersion, "Protocol Version Not supported"}
		case ConnectionRefusedIdentifierRejected:
			return &mqttErr{ConnectionRefusedIdentifierRejected, "Identifier rejected"}
		case ConnectionRefusedServerUnavailable:
			return &mqttErr{ConnectionRefusedServerUnavailable, "Server Unavailable"}
		case ConnectionRefusedUsernamePassword:
			return &mqttErr{ConnectionRefusedUsernamePassword, "Wrong Username or Password"}
		case ConnectionRefusedNotAuthorized:
			return &mqttErr{ConnectionRefusedNotAuthorized, "Not Authorised"}
		}
	} else {
		return &mqttErr{-1, "Connack Packet Not received"}
	}
	//defer c.conn.Close()
	return nil
}

func (c *Client) AwaitMessages() {
	b := make([]byte, 1)
	store := make([]byte, 0)
	var startMsg bool = false
	var i int
	for {
		select {
		case <-c.done:
			c.conn.Close()
			os.Exit(0)
			break
		case pong := <-c.readPong:
			for pong == true {
				pong = <-c.readPong
			}
		default:
			err := c.conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
			if err != nil {
				fmt.Println(err.Error())
			}
			i, _ = c.GetSrvPacket(b)
			if i > 0 {
				store = append(store, b...)
				startMsg = true
			} else {
				if startMsg {
					if len(store) > 0 {
						data, err := c.DecodePublish(store)
						if err != nil {
							fmt.Println(err.Error())
						} else {
							c.OnMessage(data)
						}
					}
					store = nil
				}
				startMsg = false
			}
		}

	}
}

func (c *Client) SendPing() {
	pkt := packet{}
	pkt.configurePingRequest()
	pingBytes := pkt.FormulateMQTTOutputData()
	d := time.Duration(c.options.KeepAlive)
	resp := make([]byte, 1)
	for {
		select {
		case <-c.done:
			c.conn.Close()
			os.Exit(0)
			break
		default:
			time.Sleep(d * time.Second)
			c.readPong <- true
			_, err := c.conn.Write(pingBytes)
			if err != nil {
				fmt.Println(err.Error())
				c.done <- 1
			}
			c.conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond)) //max read deadline to avoid missing pong response
			_, err = c.GetSrvPacket(resp)
			if resp[0] == ControlPktPingResp {
				_, err = c.conn.Read(resp)
			}
			if err != nil {
				fmt.Printf("Ping error: %s", err.Error())
			}
			c.readPong <- false

		}

	}
}
func (c *Client) Subscribe(topic string, qos QOS) error {
	pkt := packet{}
	pkt.configureSubscribePackets(topic, qos)
	c.topic = topic
	c.qos = qos
	sendBytes := pkt.FormulateMQTTOutputData()
	c.conn.Write(sendBytes)
	ctrl := make([]byte, 1)
	r, err := c.GetSrvPacket(ctrl)
	if err != nil {
		return &mqttErr{r, "Issue with getting packet"}
	}
	if int(ctrl[0]) == ControlPktSubAck {
		r, err = c.GetSrvPacket(ctrl)
		remLength := int(ctrl[0])
		data := make([]byte, remLength)
		_, err = c.GetSrvPacket(data)
		if err != nil {
			return &mqttErr{r, "Issue with getting packet"}
		}
		pktId := binary.BigEndian.Uint16(data[0:])
		if pktId != PacketIdentifier {
			return &mqttErr{int(pktId), "Packet identifier mismatch"}
		}
		if byte(qos) != data[2] {
			return &mqttErr{int(data[2]), "QOS mismatch"}
		}
	} else {
		return &mqttErr{int(ctrl[0]), "Failed to subscribe"}
	}
	go c.SendPing()
	go c.AwaitMessages()
	go c.handleSignals()
	return nil
}

//Publish send message to server
func (c *Client) Publish(message string) error {
	c.commLock.Lock()
	pkt := packet{}
	pkt.configurePublish(c.topic, message, false, c.qos)
	sendBytes := pkt.FormulateMQTTOutputData()
	c.conn.Write(sendBytes)
	c.commLock.Unlock()
	return nil
}

//DecodePublish decode message received i.e data and return MessageData type
func (c *Client) DecodePublish(data []byte) (MessageData, error) {
	var pub byte = (data[0]) & 0xf0
	msg := MessageData{}
	if pub != 0x30 {
		return MessageData{}, &mqttErr{int(pub), "Not correct packet"}
	}

	remainLen := data[1]
	topicLen := binary.BigEndian.Uint16(data[2:4])
	i := 4
	topic := make([]byte, topicLen)
	for i < int(4+topicLen) {
		topic = append(topic, data[i])
		i++
	}
	msg.Topic = string(topic)
	messageLen := int(remainLen) - (int(topicLen)) - 2
	message := make([]byte, messageLen)
	init := i
	for i < int(init+(messageLen)) {
		message = append(message, data[i])
		i++
	}
	msg.Payload = string(message)
	return msg, nil
}

//Disconnect send disconnect packet
func (c *Client) Disconnect() error {
	pkt := packet{}
	pkt.configureDisconnect()
	sendBytes := pkt.FormulateMQTTOutputData()
	c.conn.Write(sendBytes)
	return nil
}

func (c *Client) handleSignals() {
	select {
	case signal := <-c.getSignals:
		if signal == syscall.SIGINT {
			c.Disconnect()
			c.done <- 1
		}
	}
}
