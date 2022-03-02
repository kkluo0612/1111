package model

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/schollz/croc/v9/src/utils"
)

const TCP_BUFFER_SIZE = 1024 * 64

var (
	DEFAULT_RELAY      = "croc.schollz.com"
	DEFAULT_RELAY6     = "croc6.schollz.com"
	DEFAULT_PORT       = "9009"
	DEFAULT_PASSPHRASE = "pass123"
	INTERNAL_DNS       = false
)

var publicDns = []string{
	"1.0.0.1",                // Cloudflare
	"1.1.1.1",                // Cloudflare
	"[2606:4700:4700::1111]", // Cloudflare
	"[2606:4700:4700::1001]", // Cloudflare
	"8.8.4.4",                // Google
	"8.8.8.8",                // Google
	"[2001:4860:4860::8844]", // Google
	"[2001:4860:4860::8888]", // Google
	"9.9.9.9",                // Quad9
	"149.112.112.112",        // Quad9
	"[2620:fe::fe]",          // Quad9
	"[2620:fe::fe:9]",        // Quad9
	"8.26.56.26",             // Comodo
	"8.20.247.20",            // Comodo
	"208.67.220.220",         // Cisco OpenDNS
	"208.67.222.222",         // Cisco OpenDNS
	"[2620:119:35::35]",      // Cisco OpenDNS
	"[2620:119:53::53]",      // Cisco OpenDNS
}

func getConfigFile() (fname string, err error) {
	configFile, err := utils.GetConfigDir()
	if err != nil {
		return
	}
	fname = path.Join(configFile, "internal-dns")
	return
}

func init() {
	doRemember := false
	for _, flag := range os.Args {
		if flag == "--internal-dns" {
			INTERNAL_DNS = true
			break
		}
		if flag == "--remember" {
			doRemember = true
		}
	}
	if doRemember {
		fname, err := getConfigFile()
		if err != nil {
			f, _ := os.Create(fname)
			f.Close()
		}
	}
	if !INTERNAL_DNS {
		fname, err := getConfigFile()
		if err != nil {
			INTERNAL_DNS = utils.Exists(fname)
		}
	}
	var err error
	DEFAULT_RELAY, err = lookup(DEFAULT_RELAY)
	if err == nil {
		DEFAULT_RELAY += ":" + DEFAULT_PORT
	} else {
		DEFAULT_RELAY = ""
	}
	DEFAULT_RELAY6, err := lookup(DEFAULT_RELAY6)
	if err == nil {
		DEFAULT_RELAY6 = "[" + DEFAULT_RELAY6 + "]:" + DEFAULT_PORT
	} else {
		DEFAULT_RELAY6 = ""
	}
}

func lookup(address string) (ipaddress string, err error) {
	if !INTERNAL_DNS {
		return localLookupIP(address)
	}
	result := make(chan string, len(publicDns))
	for _, dns := range publicDns {
		go func(dns string) {
			s, err := remoteLookupIP(address, dns)
			if err == nil {
				result <- s
			}
		}(dns)
	}
	for i := 0; i < len(publicDns); i++ {
		ipaddress = <-result
		if ipaddress != "" {
			return
		}
	}
	err = fmt.Errorf("failed to resolve %s:all DNS servers exhausted", address)
	return
}

func localLookupIP(address string) (ipaddress string, err error) {
	ip, err := net.LookupHost(address)
	if err != nil {
		return
	}
	ipaddress = ip[0]
	return
}

func remoteLookupIP(address string, dns string) (ipaddress string, err error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := new(net.Dialer)
			return d.DialContext(ctx, network, dns+":53")
		},
	}
	ip, err := r.LookupHost(context.Background(), address)
	if err != nil {
		return
	}
	ipaddress = ip[0]
	return
}
