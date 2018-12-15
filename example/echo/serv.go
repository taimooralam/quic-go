
package main

import (
	"os"
	"bytes"
	"io"
	"log"
	"strings"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"encoding/binary"
	"fmt"
	"time"
	"math/big"

	"strconv"
	"path/filepath"

	quic "github.com/lucas-clemente/quic-go"
)

const numberOfPacket = 50000
const printingSpace = 2500


const addr = "127.0.0.1:4242"

const message = "foobar"

const log_file_name = "/home/alam/code/thesis/test_protocol/data/Log.txt"

const make_graph_file_name = "../../../../../../../thesis/test_protocol/graph.py"

//helper method to convert an int64 array to string
func SplitToString(array []float64, sep string) string {
	if len(array) == 0 {
		return ""
	}

	b := make([]string, len(array))
	for i, v := range array {
		b[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return strings.Join(b, sep)
}

type Packet struct {
	SequenceNumber int32
	TimeStamp int64
}

type Latency struct {
	Reader io.Reader
	NumberOfPackets int
	Data []float64
	Cursor int
	File *os.File
}


//create a new latency struct
func newLatencyReader(reader io.Reader, packets int) *Latency{
	//write the changes to the file
	//open the file
	path, _ := filepath.Abs(log_file_name)
	fmt.Println(path)
	//f, err := os.Create(path)
	f, err := os.OpenFile(log_file_name, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0666)
	check(err)
	return &Latency{Reader: reader, NumberOfPackets: packets, Data:make([]float64 , packets), Cursor: 0, File: f}
}

//place the packet data into the relevant place in the Data array of the latency struct
func (l *Latency) place(packet Packet) {
	l.Data[packet.SequenceNumber] = float64(time.Now().UnixNano() - packet.TimeStamp)/float64(1000000000)
	_, err := l.File.WriteString( strconv.FormatInt(int64(packet.SequenceNumber), 10) +":" + strconv.FormatFloat(l.Data[packet.SequenceNumber], 'f', -1, 64) + " ")
	check(err)
}

//Read the data as the read from stream is given
func (l *Latency) Read(p []byte) (int, error) {
	fmt.Println(l.Cursor)
	//return if the cursor equals the number of packets
	if l.Cursor == l.NumberOfPackets {
		return 0, io.EOF
	}

	// read the data
	n, err := io.ReadFull(l.Reader, p)
	if err != nil {
		return n, err
	}

	//increment the cursor
	l.Cursor = l.Cursor + 1

	//convert the data into a packet
	var packet Packet
	//json.Unmarshal(p[:n], &packet)
	if err := binary.Read(bytes.NewReader(p), binary.LittleEndian, &packet); err!= nil{
		fmt.Println("Binary.read failed: ", err)
	}

	//fmt.Print('*')
	//place the packet
	l.place(packet)

	//return the correct number of bytes Read
	return n, nil
}


func check(e error){
	if e!= nil {
		panic(e)
	}
}

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	fmt.Printf("Hello world\n")
	//go func() { log.Fatal(echoServer()) }()
	echoServer()
}

// Start a server that echos all data on the first stream opened by the client
func echoServer() error {

	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	check(err)

	sess, err := listener.Accept()
	check(err)

	stream, err := sess.AcceptStream()
	check(err)
	// Echo through the loggingWriter
	//_, err = io.Copy(loggingWriter{os.Stdout}, stream)

	//make a latency reader
	latencyReader := newLatencyReader(stream, numberOfPacket)
	defer latencyReader.File.Close()

	//first write the number to the file
	_, err = latencyReader.File.WriteString(strconv.FormatInt(numberOfPacket,10)+"\n")
	check(err)

	//make a buffer p of length 12 bytes
	p := make([]byte, 12)

	//create a for loop to read the buffer continuously
	for {
		_, err := latencyReader.Read(p)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(string(p[:n]))
	}

	//print the data in the latency reader to check the output
	for i:=0; i < len(latencyReader.Data) ; i+=printingSpace {
		var latencyms = float64(latencyReader.Data[i])/float64(1000000000)
		fmt.Printf("Seq: %d, Delta: %v ms\n",i, latencyms)
	}


	//next write the buffer to the file
	//_, err = f.WriteString(SplitToString(latencyReader.Data, ","))
	//check(err)

	//running the graph.py python program
	//graph_path, err := filepath.Abs(make_graph_file_name)
	//fmt.Println(graph_path)
	//check(err)
	//cmd := exec.Command("sudo","python", string(graph_path))
	//fmt.Println("Running the graph command and waiting for it to finish")
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//err = cmd.Run()
	//check(err)
	//fmt.Println("Output of python graph.py: \n%q", out.String())
	return err
}


// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	//fmt.Printf("Server: Got '%s'\n\n", string(b))
	fmt.Printf("\n[x]")
	//declare the empty object
	//var packet Packet
	//json.Unmarshal(b, &packet)
	//print the packet now
	//fmt.Printf("Printing the packet now - Sequence Number: %d, TimeStamp: %d, Data: %s\n", packet.SequenceNumber, packet.TimeStamp, packet.Data)
	return w.Writer.Write(b)
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
