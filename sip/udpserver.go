package sip

import (
	"io"
	"log"
	"net"
	"bufio"
	"bytes"
	"strconv"
	"net/textproto"
)

type UdpClient struct {
	src *net.UDPAddr
	srv *Server
	UserClient SipHandler
}

func (c *UdpClient)HandleMsg(msg *SipMsg) {
	if msg.IsRequest() {
		c.UserClient.HandleRequest(msg)
	} else {
		c.UserClient.HandleResponse(msg)
	}
}

func (srv *Server)serveUdp(c *net.UDPConn){
	var buf [512]byte
	log.Println("start looping")
	for {
		n, addr, err := c.ReadFromUDP(buf[0:])
		if err!=nil {
			log.Println(err)
			continue
		}
		msg := bytes.NewBuffer(buf[:n]) //simplified reading single shot
		tpr := textproto.NewReader(bufio.NewReader(msg))
		sl, err:= tpr.ReadLine()
		if err != nil {
			if err==io.EOF {
				break
			} else {
				continue 
			}
		}
		if !isSipStart(sl) {
			continue
		}
		log.Println(sl)
		sipmsg := new(SipMsg);
		sipmsg.StartLine = sl
		sipmsg.Headers, err = tpr.ReadMIMEHeader()
		if err != nil {
			log.Println(err)
			continue
		}
		if v,ok:= srv.Clients[addr.String()]; ok {
			v.HandleMsg(sipmsg)
		} else {
			tmp := &UdpClient{
				addr,
				srv,
				srv.Handler(),
			}
			srv.Clients[addr.String()] = tmp
			tmp.HandleMsg(sipmsg)
		}
	}
}

func (srv *Server)ServeUdp() error {
	log.Println("START SERVER")
	srv.Wait.Add(1)
	baseAddr := srv.BindIP+":"+strconv.Itoa(srv.UdpPort)
	udpAddr, err := net.ResolveUDPAddr("udp",baseAddr)
	if err != nil {
		log.Println(err)
	}
	c, e := net.ListenUDP("udp",udpAddr)
	if e!= nil {
		log.Println(e)
		return e
	}
	
	srv.serveUdp(c)
	srv.Wait.Done()
	return nil
}


func (srv *Server)ServeMulticastUdp() error {
	log.Println("START MULTICAST SERVER")
	srv.Wait.Add(1)
	baseAddr := srv.BindIP+":"+strconv.Itoa(srv.UdpPort)
	mcastAddr := srv.Multicast+":"+strconv.Itoa(srv.UdpPort)
	_, err := net.ResolveUDPAddr("udp",baseAddr)  // was lAddr
	mAddr, err := net.ResolveUDPAddr("udp",mcastAddr)
	if err != nil {
		log.Println(err)
	}
	m, e := net.ListenMulticastUDP("udp",nil,mAddr)
	if e!= nil {
		log.Println(e)
		return e
	}
	srv.Wait.Done()
	// _, e = net.ListenUDP("udp",lAddr) // was c
	// if e!= nil {
	// 	log.Println(e)
	// 	return e
	// }
	
	srv.serveUdp(m)

	return nil
}
