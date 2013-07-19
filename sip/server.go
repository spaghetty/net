package sip

import (
	"log"
	"net"
	"sync"
	"bytes"
	"strconv"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

type SipHandler interface {
	SubscribeRequest(*Request)
	HandleRequest(*Request)
	HandleResponse(*Response)
	GetStack() Stack
}

type ResponseWriter interface {
        // Header returns the header map that will be sent by WriteHeader.
        // Changing the header after a call to WriteHeader (or Write) has
        // no effect.
        Header() Header

        // Write writes the data to the connection as part of an HTTP reply.
        // If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
        // before writing the data.  If the Header does not contain a
        // Content-Type line, Write adds a Content-Type set to the result of passing
        // the initial 512 bytes of written data to DetectContentType.
        Write([]byte) (int, error)

        // WriteHeader sends an HTTP response header with status code.
        // If WriteHeader is not called explicitly, the first call to Write
        // will trigger an implicit WriteHeader(http.StatusOK).
        // Thus explicit calls to WriteHeader are mainly used to
        // send error codes.
        WriteHeader(int)
}

// A Server defines parameters for running an SIP server.
type Server struct {
	Wait              *sync.WaitGroup
	WriteUdp          *sync.Mutex
	BindIP            string
	Multicast         string
	TcpPort           int        // TCP address to listen on, ":http" if empty
	UdpPort           int        // Unimplemented
	TlsPort           int        // Unimplemented
	Handler           func()SipHandler    // handler to invoke, http.DefaultServeMux if nil
	Clients           map[string]EndPoint
	udpConn           *net.UDPConn
	tcpConn           *net.TCPConn
	UDPContact        EUri
}


func NewServer(BindIp string, Multicast string, TcpPort int, UdpPort int, h func()SipHandler) *Server {
	var u EUri;
	u.U.Schema = "sip"
	u.U.Host = BindIp
	u.U.Port = strconv.Itoa(UdpPort)
	u.Parameters = make(map[string]string)
	u.U.Parameters = make(map[string]string)
	return &Server{
		new(sync.WaitGroup),
		new(sync.Mutex),
		BindIp,
		Multicast,
		TcpPort,
		UdpPort,
		0,
		h,
		make(map[string]EndPoint),
		nil,
		nil,
		u,
	}
}

func (srv *Server)WriteUDP(msg SipMsg, add *net.UDPAddr) {
	buf := bytes.NewBufferString("")
	msg.Write(buf)
	srv.WriteUdp.Lock()
	i, err := srv.udpConn.WriteToUDP(buf.Bytes(), add)
	srv.WriteUdp.Unlock()
	log.Println("toh : ", i , " " , err)
}

func (srv *Server)Run() error{
	if srv.Multicast!="" {
		log.Println("TRYING FOR MULTICAST")
		srv.Wait.Add(1)
		go srv.ServeMulticastUdp()
	}
	if srv.BindIP!="" {
		if srv.TcpPort!=0 {
			srv.Wait.Add(1)
			log.Println("TRYING FOR TCP")
			// FIXME ****
		}
		if srv.UdpPort!=0 {
			srv.Wait.Add(1)
			log.Println("TRYING FOR UDP")
			go srv.ServeUdp()
		}
	}
	srv.Wait.Wait()
	return nil
}
