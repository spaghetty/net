package sip

import (
	"sync"
)

type SipHandler interface {
	HandleRequest(*SipMsg)
	HandleResponse(*SipMsg)
	//SetStack(*Stack)
}

// A Server defines parameters for running an SIP server.
type Server struct {
	Wait              *sync.WaitGroup
	BindIP            string
	Multicast         string
	TcpPort           int        // TCP address to listen on, ":http" if empty
	UdpPort           int        // Unimplemented
	TlsPort           int        // Unimplemented
	Handler           func()SipHandler    // handler to invoke, http.DefaultServeMux if nil
	Clients           map[string]SipEndPoint
}


func NewServer(BindIp string, Multicast string, TcpPort int, UdpPort int, h func()SipHandler) *Server {
	return &Server{
		new(sync.WaitGroup),
		BindIp,
		Multicast,
		TcpPort,
		UdpPort,
		0,
		h,
		make(map[string]SipEndPoint),
	}
}


func (srv *Server)Run() error{
	if srv.Multicast!="" {
		go srv.ServeMulticastUdp()
	}
	if srv.BindIP!="" {
		if srv.TcpPort!=0 {
			// run tcp server
		}
		if srv.UdpPort!=0 {
			go srv.ServeUdp()
		}
	}
	return nil
}
