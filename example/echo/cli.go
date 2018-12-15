package main

import (
	//"crypto/rand"
	//"crypto/rsa"
	"crypto/tls"
	"encoding/binary"
	"time"
	"bytes"
	//"crypto/x509"
	//"encoding/pem"
	"fmt"
	"reflect"
	//"io"
	//"log"
	//"math/big"

	quic "github.com/lucas-clemente/quic-go"
)

const addr = "127.0.0.1:4242"

const message = "foobar"

const numPackets = 50000

type Packet struct {
	SequenceNumber int32
	TimeStamp int64
}

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {

	err := clientMain()
	if err != nil {
		panic(err)
	}
}


func clientMain() error {
	session, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, nil)
	if err != nil {
		return err
	}

	stream, err := session.OpenStreamSync()
	if err != nil {
		return err
	}

	fmt.Println(reflect.TypeOf(stream))

	for i:=0; i < numPackets ; i++ {
		//create a packet
		packet := Packet{int32(i), time.Now().UnixNano()}

		//marshall the packets to bytes
		//bytes, _ := json.Marshal(packet)
		buffer := new(bytes.Buffer)

		//create an interface of the packet structure so that it is iterable
		reflection := reflect.ValueOf(packet)
		values:= make([]interface{}, reflection.NumField())
		for i:=0; i < reflection.NumField(); i++ {
			values[i] = reflection.Field(i).Interface()
		}

		//run the for loop to binary marshallize the packet from the interface of the packet
		for _, v := range values {
			err := binary.Write(buffer, binary.LittleEndian, v)
			if err != nil {
				fmt.Println("binary.Write failed: ", err)
			}
		}

		//converting the buffer in bytes to send
		bytes := buffer.Bytes()

		//optional - print the marshalled json object along with its length
		fmt.Printf("Sending the data: %x with length: %d\n", bytes, len(bytes))

		//send the bytes to the server
		_, err = stream.Write(buffer.Bytes())
		if err != nil {
			return err
		}
		fmt.Println(i)
	}
	stream.Close()

	return nil
}
