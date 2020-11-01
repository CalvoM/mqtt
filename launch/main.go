package main

import (
	"io"
	"math"
"github.com/CalvoM/mqtt"
)

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

func main() {
	client:= mqtt.Client{
		Host:"mqtt.m-vend.com",
		Port:"1883",
	}
	options:=mqtt.ConnectOptions{
		CleanSession:true,
		KeepAlive:16,
		Password:"john",
		Username:"john",
		WillFlag:true,
		WillTopic:"v/t/will",
		WillMessage:"Goodbye",
		WillQOS:2,
	}
	client.Init(&options)
	client.Connect()
}
