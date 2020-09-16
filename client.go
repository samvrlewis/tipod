package tipod

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/samvrlewis/tipod/encoding"
	"github.com/samvrlewis/tipod/tun"
)

const (
	MTU_SIZE = 150
)

type Client struct {
	dnsDomain string
	pollTime  time.Duration
	incoming  chan []byte
	resolver  *net.Resolver
	tun       *tun.Tun
	quit      chan bool
}

func NewClient(dnsDomain string, pollTime time.Duration, nameServer string, nameServerPort int, tun *tun.Tun) *Client {
	client := Client{dnsDomain: dnsDomain, pollTime: pollTime, tun: tun}
	client.incoming = make(chan []byte)
	client.quit = make(chan bool)

	if nameServer != "" {
		client.resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", net.JoinHostPort(nameServer, strconv.Itoa(nameServerPort)))
			},
		}
	} else {
		client.resolver = net.DefaultResolver
	}

	return &client
}

func (c *Client) dNSRequestLoop() {
	for {
		// Todo: Have a nicer way of sending this "I'm not sending any data" metadata
		domainPrefix := "notreal."
		var err error

		select {
		case <-c.quit:
			return
		case data := <-c.incoming:
			domainPrefix, err = encoding.ToDomainPrefix(data)
			if err != nil {
				log.Fatalln(err)
			}
		case <-time.After(c.pollTime):
			break
		}

		domainPrefix = domainPrefix + c.dnsDomain
		dataBack, err := c.resolver.LookupTXT(context.Background(), domainPrefix)

		if err != nil {
			log.Println(err)
			continue
		}

		if dataBack[0] == "" {
			continue
		}

		bytesBack, err := encoding.FromTxtData(dataBack[0])

		if err != nil {
			log.Fatalln(err)
		}

		n, err := c.tun.Write(bytesBack)

		if err != nil {
			panic(err)
		}

		if n != len(bytesBack) {
			log.Println("Didn't write all bytes")
		}
	}
}

func (c *Client) tunReadLoop() {
	packet := make([]byte, MTU_SIZE)
	for {
		// Check if we should quit before continuing
		select {
		case <-c.quit:
			return
		default:
			break
		}

		n, err := c.tun.Read(packet)
		if err != nil {
			log.Println(err)
		}
		c.incoming <- packet[:n]
	}
}

func (c *Client) Start() {
	go c.tunReadLoop()
	go c.dNSRequestLoop()
}

func (c *Client) Stop() {
	close(c.quit)
}
