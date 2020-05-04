package mqtt

//Client :
type Client struct{}

//packet
type packet struct {
	fixedHeader    []byte
	variableHeader []byte
	payLoad        []byte
}
