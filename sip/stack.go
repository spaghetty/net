package sip

import (
	//"log"
	"bytes"
	"math/rand"
)

type Dialog struct {
	CallId string
	Me     EUri
	Other  EUri
	COther EUri
}

type Stack interface {
	SetContact(c *EUri)
	SetEndPoint(EndPoint)
	SetUserInterface(SipHandler)
	HandleRequest(*Request)
	HandleResponse(*Response)
	BuildResponse(*Request, int) *Response
	BuildRequest(t MethodType, d Dialog) *Request
	GetDialog(SipMsg) Dialog
	Send(SipMsg)
}

type UserAgent struct {
	EP EndPoint
	UI SipHandler
	C  *EUri
	Tag string
}

func NewUserAgent() *UserAgent{
	u := new(UserAgent)
	u.Tag = generateTag(7)
	return u
}

func (u *UserAgent)SetContact(c *EUri){
	u.C = c
}

func (u *UserAgent)SetUserInterface(h SipHandler){
	u.UI = h
}

func (u *UserAgent)SetEndPoint(e EndPoint) {
	u.EP = e
}

func (u *UserAgent)HandleRequest(r *Request) {
	switch r.Method {
	case "SUBSCRIBE":
		u.UI.SubscribeRequest(r)
	default:
		u.UI.HandleRequest(r)
	}
}

func (u *UserAgent)GetDialog(s SipMsg) Dialog {
	return getDialog(s, u.Tag)
}

func (u *UserAgent)HandleResponse(r *Response) {
	u.UI.HandleResponse(r)
}

func (u *UserAgent) BuildResponse(r *Request, status int) *Response{
	res := new(Response)
        res.Proto = "SIP/2.0"
        res.ProtoMajor = 2
        res.ProtoMinor = 0
	res.Header = r.Header
	res.Header.Del("Event")
	res.Header.Del("Accept")
        res.Request = r
        res.StatusCode = status
        res.Status = StatusText(status)
	res.Header["Contact"][0] = u.C.String()
	to := ParseEUri(res.Header["To"][0])
	if _,ok := to.Parameters["tag"]; !ok {
		to.Parameters["tag"] = u.Tag
		res.Header["To"][0] = to.String()
	}
        switch status {
        case StatusOK:
                break
        default:
        }
        return res
}

func (u *UserAgent)BuildRequest(t MethodType, d Dialog) *Request{
	res,_ := NewRequest("NOTIFY", d.COther.U.String(),nil)
	res.Header.Add("Call-Id", d.CallId)
	res.Header.Add("From",d.Me.String())
	res.Header.Add("To",d.Other.String())
	return res
}

func (u *UserAgent)Send(msg SipMsg) {
	u.EP.SendMsg(msg)
}

func generateTag(size int) string{
	r := bytes.NewBufferString("")
	for i:=0; i < size; i++ {
		r.WriteByte(TagCharset[rand.Intn(len(TagCharset))])
	}
	return r.String()
}
