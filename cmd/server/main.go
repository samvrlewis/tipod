package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/samvrlewis/tipod/encoding"
	"github.com/samvrlewis/tipod/tun"
	"golang.org/x/net/ipv4"
)

var tunnel *tun.Tun
var waitingData [][]byte
var lock sync.Mutex

const (
	MTU_SIZE = 150
	PREFIX   = ".samlewis.me"
)

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeTXT:
			log.Println("Request to: ", q.Name)
			lastInd := strings.LastIndex(q.Name, PREFIX)
			log.Println("Data is: ", q.Name[:lastInd])

			data, err := encoding.FromDomainPrefix(q.Name[:lastInd])

			if err != nil {
				log.Fatalln(err)
			}

			if q.Name[:lastInd] != "notreal" {
				log.Println("Writing to tun: ", data)
				n, err := tunnel.Write(data)
				if err != nil {
					log.Fatalln(err)
				}

				if n != len(data) {
					log.Fatalln("Didn't write all data")
				}
			}

			t := &dns.TXT{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
				Txt: []string{""},
			}

			lock.Lock()
			if len(waitingData) > 0 {
				dataToSend, err := encoding.ToTxtData(waitingData[0])

				if err == nil {
					t.Txt = []string{dataToSend}
				}

				waitingData = waitingData[1:]
			}
			lock.Unlock()
			m.Answer = append(m.Answer, t)
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	fromIP, fromPort, err := net.SplitHostPort(w.RemoteAddr().String())

	if err != nil {
		log.Println("Error decoding from IP: ", err)
		return
	}

	fmt.Println("Request from ", fromIP, fromPort)

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {
	var err error
	tunnel, err = tun.NewTun("tipod_tun")

	if err != nil {
		log.Fatalln("err ", err)
	}

	if err := tunnel.SetLinkUp(); err != nil {
		log.Fatalln("Error setting TUN link up: ", err)
	}

	if err := tunnel.SetNetwork("192.168.50.1/24"); err != nil {
		log.Fatalln("Error setting network: ", err)
	}

	if err := tunnel.SetMtu(MTU_SIZE); err != nil {
		log.Fatalln("Error setting network: ", err)
	}

	waitingData = make([][]byte, 0)

	go func() {
		packet := make([]byte, MTU_SIZE)

		for {
			n, err := tunnel.Read(packet)
			if err != nil {
				log.Println(err)
			}
			log.Printf("Received from TUN: % x\n", packet[:n])

			header, _ := ipv4.ParseHeader(packet[:n])

			log.Printf("isTCP: %v, header: %s", header.Protocol == 6, header)
			log.Println("From IP: ", header.Src.String())
			log.Println("To IP: ", header.Dst.String())

			lock.Lock()
			waitingData = append(waitingData, packet[:n])
			lock.Unlock()

		}
	}()

	dns.HandleFunc(".", handleDnsRequest)

	server := &dns.Server{Addr: ":54", Net: "udp"}
	err = server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
