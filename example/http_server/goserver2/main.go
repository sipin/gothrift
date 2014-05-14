package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dworld/gothrift/example/http_server/test"
	"github.com/dworld/gothrift/thrift"
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

	http.Handle("/api", thrift.NewHTTPHandler(processor, protocolFactory, protocolFactory, nil))
	log.Fatal(http.ListenAndServe(":19090", nil))
}
