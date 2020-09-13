package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/samvrlewis/tipod/encoding"
	"github.com/samvrlewis/tipod/tun"
	"golang.org/x/net/ipv4"
)

const (
	MTU_SIZE = 150
)

var lock sync.Mutex

func main() {
	tunName := flag.String("tun-name", "tun", "tun name")
	tunIp := flag.String("tun-ip", "192.168.50.2/24", "tun ip")
	flag.Parse()
	tun, err := tun.NewTun(*tunName)

	if err != nil {
		log.Fatalln("err ", err)
	}

	if err := tun.SetLinkUp(); err != nil {
		log.Fatalln("Error setting TUN link up: ", err)
	}

	if err := tun.SetNetwork(*tunIp); err != nil {
		log.Fatalln("Error setting network: ", err)
	}

	if err := tun.SetMtu(MTU_SIZE); err != nil {
		log.Fatalln("Error setting network: ", err)
	}

	var resolver *net.Resolver
	nameserver := "10.1.1.1"

	// don't think I need to do this - just use the system dns
	if nameserver != "" {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", net.JoinHostPort(nameserver, "54"))
			},
		}
	} else {
		resolver = net.DefaultResolver
	}

	toDns := make(chan []byte)

	go func() {
		for {
			domainPrefix := "notreal."

			select {
			case data := <-toDns:
				log.Println("Got data: ", data)
				domainPrefix, err = encoding.ToDomainPrefix(data)
				if err != nil {
					log.Fatalln(err)
				}
			case <-time.After(20 * time.Millisecond):
				fmt.Println("timeout 1")
			}

			domainPrefix = domainPrefix + "samlewis.me"
			log.Println("Sending request: ", domainPrefix)

			dataBack, err := resolver.LookupTXT(context.Background(), domainPrefix)

			if err != nil {
				log.Println(err)
				continue
			}

			if dataBack[0] == "" {
				log.Println("Empty!")
				continue
			}

			bytesBack, err := encoding.FromTxtData(dataBack[0])

			if err != nil {
				log.Fatalln(err)
			}

			log.Println("Writing to tun: ", bytesBack)
			tun.Write(bytesBack)
		}
	}()

	go func() {
		packet := make([]byte, MTU_SIZE)

		for {
			n, err := tun.Read(packet)
			if err != nil {
				log.Println(err)
			}
			log.Printf("Received from TUN: % x\n", packet[:n])

			header, _ := ipv4.ParseHeader(packet[:n])

			log.Printf("isTCP: %v, header: %s", header.Protocol == 6, header)
			log.Println("From IP: ", header.Src.String())
			log.Println("To IP: ", header.Dst.String())

			toDns <- packet[:n]
		}
	}()

	for {

	}
}
