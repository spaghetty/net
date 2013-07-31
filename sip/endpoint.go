package sip

import (
	"net"
	"log"
)

type EndPoint interface {
	HandleMsg(SipMsg)
	GetStack() Stack
	SetStack(Stack)
	SendMsg(SipMsg)
	GetContact() EUri
	GetServer() *Server
	Copy() EndPoint
}

type UdpClient struct {
	Src *net.UDPAddr
	Srv *Server
	Stk Stack
	C   *EUri
}

func (c *UdpClient)GetServer() *Server {
	return c.Srv
}

func (c *UdpClient)GetStack() Stack {
	return c.Stk
}

func (c *UdpClient)Copy() EndPoint {
	t := new(UdpClient)
	t.Src = c.Src
	t.Srv = c.Srv
	t.C = c.C
	t.Stk = c.Stk.Copy()
	return t
}

func (c *UdpClient)SetStack(s Stack) {
	c.Stk = s
}

func (c *UdpClient)HandleMsg(msg SipMsg) {
	if c.C==nil {
		c.C = ParseEUri(msg.GetHeader().Get("Contact"))
		log.Println(c.C)
	}
	switch msg.(type) {
	case *Request:
		c.Stk.HandleRequest(msg.(*Request))
	case *Response:
		c.Stk.HandleResponse(msg.(*Response))
	}
}

func (c *UdpClient)GetContact() EUri {
	return *(c.C)
}

func (c *UdpClient)SendMsg(m SipMsg) {
	c.Srv.WriteUDP(m, c.Src)
}
