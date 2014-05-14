package main

import (
	"fmt"
	"os"

	"github.com/dworld/gothrift/example/http_transport/test"
	"github.com/dworld/gothrift/thrift"
)

func main() {
	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	// transport, err := thrift.NewTSocket(net.JoinHostPort("127.0.0.1", "19090"))
	// for goserver
	// transport, err := thrift.NewTHttpPostClient("http://127.0.0.1:19090/")
	// for goserver and goserver2
	transport, err := thrift.NewTHttpPostClient("http://127.0.0.1:19090/api")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error resolving address:", err)
		os.Exit(1)
	}
	defer transport.Close()

	useTransport := transportFactory.GetTransport(transport)
	client := test.NewTestClientFactory(useTransport, protocolFactory)
	if err := transport.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "Error opening socket to 127.0.0.1:19090", " ", err)
		os.Exit(1)
	}
	r, e := client.Hello("world")
	if e != nil {
		fmt.Fprintln(os.Stderr, "Error ", err)
		os.Exit(1)
	}
	fmt.Println(r)
	fmt.Println("OK")
}
