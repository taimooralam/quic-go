package main

import (
	"bufio"
	//"fmt"
	"log"
	"os"
	//"encoding/base64"
	//"encoding/binary"
	"encoding/gob"
	"time"
	//"strings"
	"bytes"
	"crypto/tls"
	//"reflect"
	//"io/ioutil"
	//"syscall"
	quic "github.com/lucas-clemente/quic-go"
)

const addr = "127.0.0.1:4242"

type Packet struct {
	SequenceNumber int32
	TimeStamp int64
	Data string
}

var pipeFile = "/tmp/pipe.log"

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {

        err := clientMain()
        if err != nil {
                panic(err)
        }
}


func clientMain() error {
	//open the session
        session, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, nil)
        if err != nil {
		log.Fatal(err)
                return err
        }

	

	//open the stream
        stream, err := session.OpenStreamSync()
        if err != nil {
		log.Fatal(err)
                return err
        }


	log.Println("open a named pipe file for read.")
        file, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
        if err != nil {
                log.Fatal("Open named pipe file error:", err)
		return err
        }

        reader := bufio.NewReader(file)
	i := 0

	for {
                line, err := reader.ReadString('~')
                //line, err := ioutil.ReadAll(reader)
                if err == nil {
                        //Separate the encoded data from the last separator character
			//var encoded_data = strings.TrimRight(line, " ")
			var encoded_data = line
			log.Println(len(line))

			//test to decode the base64 string
                        /*data, err := base64.StdEncoding.DecodeString(encoded_data)
                        if err != nil {
                                panic(err)
                        }*/

			//create the packet
			packet := Packet{int32(i), time.Now().UnixNano(),encoded_data}

                	//marshall the packets to bytes
                	encbuffer := new(bytes.Buffer)

			//encode the in the buffer using gob encoding
			err := gob.NewEncoder(encbuffer).Encode(packet)
			if err != nil {
				log.Fatal(err)
			}

			//converting the buffer in bytes to send
			encbuffer.WriteByte(byte('`'))
			bytes_data := encbuffer.Bytes()


			//send the data
			_, err = stream.Write(bytes_data)
			if err != nil{
				log.Fatal(err)
				return err
			}else{
				log.Print("w")
			}
                }else{
                        //fmt.Println(buffer.String())
                        log.Println("End of pipe:" + err.Error())
                        break
                }
		i++
        }
	stream.Close()

        return nil
}


