package main

import (
	"fmt"
	"os"

	"github.com/sipin/gothrift/example/http_transport/test"
	"github.com/sipin/gothrift/thrift"
)

const (
	NetworkAddr = "127.0.0.1:19090"
)

type TestImpl struct {
}

func (t *TestImpl) Hello(name string) (string, error) {
	fmt.Printf("-->Hello: %s\n", name)
	return "hello " + name, nil
}

func main() {
	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transport, err := thrift.NewTServerHttp(NetworkAddr)
	// transport, err := thrift.NewTServerSocket(NetworkAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testImpl := &TestImpl{}
	processor := test.NewTestProcessor(testImpl)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	err = server.Serve()
	if err != nil {
		fmt.Println(err)
	}
}
