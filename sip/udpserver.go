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
	srv.checkPendingClient()
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
		var smsg SipMsg
		if isRequest(sl) {
			smsg,err = ReadRequest(bufio.NewReader(bytes.NewBuffer(buf[:n])))
			if err!=nil {
				log.Println(err)
				continue
			}
		} else {
			smsg,err = ReadResponse(bufio.NewReader(bytes.NewBuffer(buf[:n])),nil)
			if err!=nil {
				log.Println(err)
				continue
			}
		}
		var ep EndPoint
		var ok bool
		var id,ide string
		id,ide = smsg.GetID(),smsg.GetEarlyID()
		if ep,ok = srv.Clients[id]; !ok {
			if ep,ok = srv.Clients[ide]; !ok {
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
				// here we need 	x.GetStack().SetIdentity(id)
				// and dialog will be builded when message was routed to stack
				ep.SetStack(x.GetStack())
				srv.Clients[ide] = ep
			} else {
				if id!="" {
					srv.Clients[id] = srv.Clients[ide].Copy()
					ep = srv.Clients[id]
				}
			}
		}
		ep.HandleMsg(smsg)

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
	if e!= nil {
		log.Println(e)
		return e
	}
	srv.udpConn = c
	srv.serveUdp(c)
	srv.Wait.Done()
	return nil
}

// contact new client is a trouble because we are in early dialog
// func (srv *Server)NewUdpEndPoint(addr *net.UDPAddr) error {
// 	ep := &UdpClient{
// 		addr,
// 		srv,
// 		nil,
// 		nil,
// 	}
// 	x := srv.Handler()
// 	x.GetStack().SetEndPoint(ep)
// 	x.GetStack().SetUserInterface(x)
// 	x.GetStack().SetContact(&(srv.UDPContact))
// 	ep.SetStack(x.GetStack())
// 	srv.Clients[addr.String()] = ep
// 	if srv.udpConn!=nil {
// 		x.GetStack().Ready()
// 	}
// 	return nil
// }

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


func (srv *Server)buildNewUdpConnection(endpoint string, id Identity) Stack {
	url := ParseEUri(endpoint)
	var addr *net.UDPAddr
	var err error;
	if p := id.GetProxy(); p!="" {
		addr,_ = net.ResolveUDPAddr("udp",p)
	} else {
		if url.U.Port=="" {
			addr,err = net.ResolveUDPAddr("udp",url.U.Host+":5060")
			if err!=nil {
				log.Println("merda", err)
			}
		} else {
			addr,_ = net.ResolveUDPAddr("udp",url.U.Host+":"+url.U.Port)
		}
	}
	ep := &UdpClient{
		addr,
		srv,
		nil,
		url,
	}
	x := srv.Handler()
	x.GetStack().SetEndPoint(ep)
	x.GetStack().SetUserInterface(x)
	x.GetStack().SetContact(&(srv.UDPContact))
	x.GetStack().SetIdentity(id)
	x.GetStack().NewDialog(endpoint)
	ep.SetStack(x.GetStack())
	// we need the key earlygetid
	cdi := x.GetStack().GetDialog()
	srv.Clients[cdi.CallId+"-"+cdi.Me.Parameters["tag"]] = ep
	return x.GetStack()
}

func (srv *Server)checkPendingClient() {
	for _,c:= range(srv.Clients){
		c.GetStack().Ready()
	}
}
