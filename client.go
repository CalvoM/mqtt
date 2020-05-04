package mqtt

//Client :
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
