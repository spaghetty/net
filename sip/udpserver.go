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

func (srv *Server)serveUdp(c *net.UDPConn){
	var buf [4096]byte
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
		var ep EndPoint
		var ok bool
		if ep,ok = srv.Clients[addr.String()]; !ok {
			ep = &UdpClient{
				addr,
				srv,
				nil,
				nil,
			}
			x := srv.Handler()
			x.GetStack().SetEndPoint(ep)
			x.GetStack().SetUserInterface(x)
			x.GetStack().SetContact(&(srv.UDPContact))
			ep.SetStack(x.GetStack())
			srv.Clients[addr.String()] = ep
		}
		if isRequest(sl) {
			smsg,err := ReadRequest(bufio.NewReader(bytes.NewBuffer(buf[:n])))
			if err!=nil {
				log.Println(err)
				continue
			}
			ep.HandleMsg(smsg)
		} else {
			smsg,_ := ReadResponse(bufio.NewReader(bytes.NewBuffer(buf[:n])),nil)
			if err!=nil {
				log.Println(err)
				continue
			}
			ep.HandleMsg(smsg)
		}

	}
}

func (srv *Server)ServeUdp() error {
	log.Println("START SERVER")
	baseAddr := srv.BindIP+":"+strconv.Itoa(srv.UdpPort)
	udpAddr, err := net.ResolveUDPAddr("udp",baseAddr)
	if err != nil {
		log.Println(err)
	}
	c, e := net.ListenUDP("udp",udpAddr)
	srv.udpConn = c;
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
	mcastAddr := srv.Multicast+":"+strconv.Itoa(srv.UdpPort)
	mAddr, err := net.ResolveUDPAddr("udp",mcastAddr)
	if err != nil {
		log.Println(err)
	}
	m, e := net.ListenMulticastUDP("udp",nil,mAddr)
	if e!= nil {
		log.Println(e)
		return e
	}
	srv.udpConn = m
	srv.serveUdp(m)
	srv.Wait.Done()
	return nil
}
