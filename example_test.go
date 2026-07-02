package grout_test

import (
	"fmt"
	"log"

	"github.com/zeeke/grout-go"
)

func ExampleConnect() {
	client, err := grout.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Printf("API version: %d\n", client.APIVersion())
	fmt.Printf("Server: %s\n", client.ServerVersion())
}

func ExampleConnect_withSocketPath() {
	client, err := grout.Connect(grout.WithSocketPath("/tmp/grout.sock"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
}

func ExampleParseIP4() {
	addr, err := grout.ParseIP4("10.0.0.1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(addr)
	// Output: 10.0.0.1
}

func ExampleParseIP4Net() {
	net, err := grout.ParseIP4Net("192.168.1.0/24")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(net)
	// Output: 192.168.1.0/24
}

func ExampleParseEtherAddr() {
	mac, err := grout.ParseEtherAddr("aa:bb:cc:dd:ee:ff")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mac)
	// Output: aa:bb:cc:dd:ee:ff
}
