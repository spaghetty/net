package sip

import (
	"net"
)

type EndPoint interface {
	HandleMsg(SipMsg)
	GetStack() Stack
	SetStack(Stack)
	SendMsg(SipMsg)
}

type UdpClient struct {
	Src *net.UDPAddr
	Srv *Server
	Stk Stack
}

func (c *UdpClient)GetStack() Stack {
	return c.Stk
}

func (c *UdpClient)SetStack(s Stack) {
	c.Stk = s
}

func (c *UdpClient)HandleMsg(msg SipMsg) {
	switch msg.(type) {
	case *Request:
		c.Stk.HandleRequest(msg.(*Request))
	case *Response:
		c.Stk.HandleResponse(msg.(*Response))
	}
}

func (c *UdpClient)SendMsg(m SipMsg) {
	c.Srv.WriteUDP(m, c.Src)
}
