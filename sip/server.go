package sip

import (
	"io"
	"net"
	"fmt"
	"log"
	"sync"
	"bytes"
	"bufio"
	"strings"
	"strconv"
	"net/textproto"
	"github.com/spaghetty/sip_parser"
)


type SipMsg sipparser.SipMsg

const noLimit int64 = (1 << 63) - 1

type SipHandler interface {
	Serve(*SipMsg)
	SetStack(*Stack)
}


type conn struct {
	remoteAddr string
	server     *Server
	rwc        net.Conn
	sr         liveSwitchReader     // where the LimitReader reads from; usually the rwc
	lr         *io.LimitedReader    // io.LimitReader(sr)
	buf        *bufio.ReadWriter    // buffered(lr,rwc), reading from bufio->limitReader->sr->rwc
	bufswr     *switchReader        // the *switchReader io.Reader source of buf
	bufsww     *switchWriter        // the *switchWriter io.Writer dest of buf
	ex         sync.Mutex
	calls      map[string]*Stack
}

// A Server defines parameters for running an SIP server.
type Server struct {
	BindIP            string
	TcpPort           int        // TCP address to listen on, ":http" if empty
	UdpPort           int        // Unimplemented
	TlsPort           int        // Unimplemented
	Handler           func()SipHandler    // handler to invoke, http.DefaultServeMux if nil
}

type switchReader struct {
	io.Reader
}

type switchWriter struct {
	io.Writer
}
// A liveSwitchReader is a switchReader that's safe for concurrent
// reads and switches, if its mutex is held.
type liveSwitchReader struct {
	sync.Mutex
	r io.Reader
}

func (sr *liveSwitchReader) Read(p []byte) (n int, err error) {
	sr.Lock()
	r := sr.r
	sr.Unlock()
	return r.Read(p)
}

func newTextprotoReader(br *bufio.Reader) *textproto.Reader {
	return textproto.NewReader(br)	
}

func isSipStart(line string) bool {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return false
	}
	s2 += s1 + 1
	return (line[:s1]=="SIP/2.0" || line[s2+1:]=="SIP/2.0")
}

func (c *conn) serve() {
	for {
		c.lr.N = 4096
		tp := newTextprotoReader(c.buf.Reader)
		var s string
		var err error
		if s, err = tp.ReadLine(); err != nil {
			if err==io.EOF {
				break
			} else {
				log.Println("Troubles reading the request line "+ err.Error())
				continue
			}
		}
		if !isSipStart(s) {
			log.Println("trash->"+s)
			continue
		}
		log.Println(s)
		mimeHeader, err := tp.ReadMIMEHeader()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(mimeHeader["Content-Length"])
		totalBody, err := strconv.Atoi(mimeHeader["Content-Length"][0])
		if err != nil {
			log.Println(err)
			continue
		}
		var body bytes.Buffer
		tmp := make([]byte, 1024)
		for body.Len()<totalBody {
			log.Println(body.Len())
			i, _ := c.buf.Reader.Read(tmp)
			body.Write(tmp[:i])
		}
		msg := bytes.NewBufferString(s)
		for k,v := range mimeHeader {
			fmt.Fprintf(msg, "%s: %s\r\n", k, v[0])
		}
		msg.WriteString("\r\n")
		msg.WriteString(body.String())
		msg.WriteString("\r\n")
		var myMsg *SipMsg = (*SipMsg)(sipparser.ParseMsg(msg.String()))
		//log.Println(msg)
		var v *Stack
		var ok bool;
		if v, ok = c.calls[myMsg.CallId]; !ok {
			v = NewStack(c, myMsg.CallId, c.server.Handler())
			c.calls[myMsg.CallId] = v
			v.Handler.SetStack(v)
		}
		
		// FIXIT: this create a limitation 1 handler (==call) per connection 
		// this works but can't be wrong.
		v.Handler.Serve(myMsg)
	}
}


func NewServer(BindIp string, TcpPort int, h func()SipHandler) *Server {
	return &Server{
		BindIp,
		TcpPort,
		0,
		0,
		h,
	}
}

func (srv *Server) checkImplementationState() bool {
	if( srv.TcpPort == 0) {
		log.Println("you must define a tcp port for now is the only protocol defined")
		return false
	}
	if( srv.UdpPort != 0 || srv.TlsPort != 0) {
		log.Println("at this time just tcp is supported, udp and tls will be ignored")
		return true
	}
	return true
}

func (srv *Server) ListenAndServe() error {
	if(!srv.checkImplementationState()) {
		log.Fatal("change setup")
	}
	tcpAddr := srv.BindIP+":"+strconv.Itoa(srv.TcpPort)
	l, e := net.Listen("tcp",tcpAddr)
	if e!= nil {
		return e
	}
	return srv.Serve(l)
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	for {
		rw, e := l.Accept()
		if e != nil {
			log.Println(e)
			continue
		}
		c, err := srv.newConn(rw)
		if err != nil {
			continue
		}
		go c.serve()
	}
	return nil
}

func (srv *Server) newConn(rwc net.Conn) (c *conn, err error) {
	c = new(conn)
	c.remoteAddr = rwc.RemoteAddr().String()
	c.server = srv
	c.rwc = rwc
	c.sr = liveSwitchReader{r: c.rwc}
	c.lr = io.LimitReader(&c.sr, noLimit).(*io.LimitedReader)
	sr := &switchReader{c.lr}
	sw := &switchWriter{c.rwc}
	br := bufio.NewReader(sr)
	bw := bufio.NewWriterSize(sw, 4<<10)
	c.buf = bufio.NewReadWriter(br, bw)
	c.bufswr = sr
	c.bufsww = sw
	c.calls = make(map[string]*Stack)
	return c, nil
}