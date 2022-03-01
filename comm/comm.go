package comm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/schollz/croc/v9/src/utils"
	log "github.com/schollz/logger"
	"golang.org/x/net/proxy"
)

var Socks5Proxy = ""

var MAGIC_BYTES = []byte("croc")

type Comm struct {
	connection net.Conn
}

func NewConnection(address string, timelimit ...time.Duration) (c *Comm, err error) {
	tlimit := 30 * time.Second
	if len(timelimit) > 0 {
		tlimit = timelimit[0]
	}

	var connection net.Conn
	if Socks5Proxy != "" && !utils.IsLocalIP(address) {
		var dialer proxy.Dialer
		if !strings.Contains(Socks5Proxy, `://`) {
			Socks5Proxy = `socks5://` + Socks5Proxy
		}
		socks5ProxyURL, urlParseError := url.Parse(Socks5Proxy)
		if urlParseError != nil {
			err = fmt.Errorf("Unable to parse socks proxy url: %v", urlParseError)
			log.Debug(err)
			return
		}
		dialer, err = proxy.FromURL(socks5ProxyURL, proxy.Direct)
		if err != nil {
			err = fmt.Errorf("proxy fialed:%w", err)
			log.Debug(err)
			return
		}
		log.Debug("dialing with dialer.Dial")
		connection, err = dialer.Dial("tcp", address)
	} else {
		log.Debugf("dialing ro %s with timelimit %s", address, tlimit)
		connection, err = net.DialTimeout("tcp", address, tlimit)
	}
	if err != nil {
		err = fmt.Errorf("comm.NewConnection fialed:%w", err)
		log.Debug(err)
		return
	}
	c = New(connection)
	log.Debugf("connected to '%s'", address)
	return
}

func New(c net.Conn) *Comm {
	if err := c.SetReadDeadline(time.Now().Add(3 * time.Hour)); err != nil {
		log.Warnf("erroe setting read deadline:%v", err)
	}
	if err := c.SetDeadline(time.Now().Add(3 * time.Hour)); err != nil {
		log.Warnf("error setting overall deadline:%v", err)
	}
	if err := c.SetWriteDeadline(time.Now().Add(3 * time.Hour)); err != nil {
		log.Errorf("error setting write deadline:%v", err)
	}
	comm := new(Comm)
	comm.connection = c
	return comm
}

func (c *Comm) Connection() net.Conn {
	return c.connection
}

func (c *Comm) Close() {
	if err := c.connection.Close(); err != nil {
		log.Warnf("error close connection:%v", err)
	}
}

func (c *Comm) Write(b []byte) (n int, err error) {
	header := new(bytes.Buffer)
	err = binary.Write(header, binary.LittleEndian, uint32(len(b)))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	tmpCopy := append(header.Bytes(), b...)
	tmpCopy = append(MAGIC_BYTES, tmpCopy...)
	n, err = c.connection.Write(tmpCopy)
	if err != nil {
		err = fmt.Errorf("connection.Write fialed:%w", err)
		return
	}
	if n != len(tmpCopy) {
		err = fmt.Errorf("wanted to write %d but wrote %d", len(b), n)
		return
	}
	return
}

func (c *Comm) Read() (buf []byte, numBytes int, bs []byte, err error) {
	if err := c.connection.SetReadDeadline(time.Now().Add(3 * time.Hour)); err != nil {
		log.Warnf("error setting read deadline:%v", err)
	}
	defer c.connection.SetDeadline(time.Time{})
	header := make([]byte, 4)
	_, err = io.ReadFull(c.connection, header)
	if err != nil {
		log.Debug("initial read error:%v", err)
		return
	}
	if !bytes.Equal(header, MAGIC_BYTES) {
		err = fmt.Errorf("initial bytes are not magic:%x", header)
		return
	}
	header = make([]byte, 4)
	_, err = io.ReadFull(c.connection, header)
	if err != nil {
		log.Debugf("initial read error:%v", err)
		return
	}
	var numBytesUint32 uint32
	rbuf := bytes.NewReader(header)
	err = binary.Read(rbuf, binary.LittleEndian, &numBytesUint32)
	if err != nil {
		err = fmt.Errorf("binary.Read fialed:%w", err)
		log.Debug(err.Error())
		return
	}
	numBytes = int(numBytesUint32)
	if err := c.connection.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Warnf("error setting read deadline:%v", err)
	}
	buf = make([]byte, numBytes)
	_, err = io.ReadFull(c.connection, buf)
	if err != nil {
		log.Debugf("consecutive read error:%v", err)
		return
	}
	return
}

func (c *Comm) Send(message []byte) (err error) {
	_, err = c.Write(message)
	return
}

func (c *Comm) Recevie() (b []byte, err error) {
	b, _, _, err = c.Read()
	return
}
