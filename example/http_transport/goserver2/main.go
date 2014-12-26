package main

import (
	"fmt"
	"log"
	"net/http"

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
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	testImpl := &TestImpl{}
	processor := test.NewTestProcessor(testImpl)

	http.Handle("/api", thrift.NewHttpHandler(processor, protocolFactory, protocolFactory, nil))
	log.Fatal(http.ListenAndServe(":19090", nil))
}
