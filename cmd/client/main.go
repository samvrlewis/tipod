package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/samvrlewis/tipod"
	"github.com/samvrlewis/tipod/tun"
)

func main() {
	tunName := flag.String("tun-name", "tun", "The name to assign to the tunnel")
	tunIP := flag.String("tun-ip", "192.168.50.2/24", "The IP address to assign to the tunnel")
	domainName := flag.String("domain-name", "samlewis.me", "Domain name to use for queries")
	dnsIP := flag.String("dns-ip", "10.1.1.1", "The DNS server to connect to")
	dnsPort := flag.Int("dns-port", 54, "Port of the DNS server to connect to")
	pollTimeMillis := flag.Int("poll-time", 50, "How frequently (in ms) to poll the DNS server")

	flag.Parse()
	tun, err := tun.NewTun(*tunName)

	if err != nil {
		log.Fatalln("err ", err)
	}

	if err := tun.SetLinkUp(); err != nil {
		log.Fatalln("Error setting TUN link up: ", err)
	}

	if err := tun.SetNetwork(*tunIP); err != nil {
		log.Fatalln("Error setting network: ", err)
	}

	if err := tun.SetMtu(tipod.MTU_SIZE); err != nil {
		log.Fatalln("Error setting network: ", err)
	}

	client := tipod.NewClient(*domainName, time.Millisecond*time.Duration(*pollTimeMillis), *dnsIP, *dnsPort, tun)

	go client.Start()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		log.Println("Shutting down")
		client.Stop()
	}
}
