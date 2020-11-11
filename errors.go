package mqtt

import "fmt"

type mqttErr struct{
	value int
	msg string
}
func (e *mqttErr) Error() string{
	return fmt.Sprintf("\r\nErr Value: %x\r\nErr Message: %s",e.value,e.msg)
}