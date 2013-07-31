package sip

import (
	"log"
	"bytes"
	"errors"
	"strconv"
	"math/rand"
)

type Dialog struct {
	CallId string
	Me     EUri
	Other  EUri
}

type Stack interface {
	SetContact(c *EUri)
	SetEndPoint(EndPoint)
	SetUserInterface(SipHandler)
	SetIdentity(Identity)
	HandleRequest(*Request)
	HandleResponse(*Response)
	BuildResponse(*Request, int) *Response
	BuildRequest(MethodType, ...string) (*Request,error)
	GetDialog() Dialog
	NewDialog(string)
	Send(SipMsg)
	Ready()
	Copy() Stack
}

type UserAgent struct {
	EP EndPoint
	UI SipHandler
	I  Identity
	C  *EUri
	Tag string
	MForwards string
	D  *Dialog
}

func NewUserAgent() *UserAgent{
	u := new(UserAgent)
	u.Tag = generateTag(7)
	u.MForwards = strconv.Itoa(70)
	return u
}

func (u *UserAgent)Copy() Stack {
	s := u.EP.GetServer().Handler()
	x := s.GetStack().(*UserAgent)
	x.Tag = u.Tag
	x.MForwards = u.MForwards
	x.C = u.C
	x.I = u.I
	x.D = u.D
	return x
}

func (u *UserAgent)SetIdentity(i Identity) {
	u.I = i
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
	if u.D==nil {
		tmp := getDialog(r, u.Tag)
		u.D = &tmp
	}
	switch r.Method {
	case "SUBSCRIBE":
		u.UI.SubscribeRequest(r)
	default:
		u.UI.HandleRequest(r)
	}
}

//getDialog(s, u.Tag)

func (u *UserAgent)GetDialog() Dialog {
	return *(u.D)
}

func (u *UserAgent)HandleResponse(r *Response) {
	if u.D==nil {
		tmp := getDialog(r, u.Tag)
		u.D = &tmp
	}
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

func (u *UserAgent)BuildRequest(t MethodType, uri ...string) (*Request, error) {
	var rc *EUri
	
	switch len(uri) {
	case 1:
		rc = ParseEUri(uri[0])
	case 0:
		tmprc := u.EP.GetContact()
		if tmprc.IsEmpty() {
			return nil, errors.New("No such url for request")
		}
		rc = &tmprc
	default:
		return nil, errors.New("too mutch parameters")
	}
	tmp,_ := u.I.GetEUri()
	var res *Request
	switch t { 
	case NOTIFY:
		res,_ = NewRequest("NOTIFY", rc.U.String(),nil)
		res.Header.Add("Cseq", "3 NOTIFY")
		res.Header.Add("Contact", u.C.String())
		res.Header.Add("Call-Id", u.D.CallId)
		res.Header.Add("From", u.D.Me.String())
		res.Header.Add("To", u.D.Other.String())
		res.Header.Add("Max-Forwards",u.MForwards)
	case REGISTER:
		res,_ = NewRequest("REGISTER", rc.U.String(),nil)
		res.Header.Add("Cseq", "1 REGISTER")
		res.Header.Add("Contact", u.C.String())
		res.Header.Add("Call-Id", u.D.CallId)
		res.Header.Add("From", tmp.String())
		res.Header.Add("To", u.D.Other.String())
		res.Header.Add("Max-Forwards",u.MForwards)
	} 
	return res, nil
}

func (u *UserAgent)Send(msg SipMsg) {
	u.EP.SendMsg(msg)
}

func generateCallId() string {
	return generateTag(15)
}

func generateTag(size int) string{
	r := bytes.NewBufferString("")
	for i:=0; i < size; i++ {
		r.WriteByte(TagCharset[rand.Intn(len(TagCharset))])
	}
	return r.String()
}

func (u *UserAgent)Ready() {
	log.Println("READY")
	u.UI.Init()
}

func (u *UserAgent)NewDialog(other string) {
	tmp := newDialog(u.C.U, u.Tag, other)
	u.D = &tmp
}

func newDialog(me Uri, mytag string , other string) (d Dialog) {
	d.Me = EUri{
		"",
		me,
		make(map[string]string),
	}
	d.Me.Parameters["tag"] = mytag
	d.CallId = generateCallId()
	d.Other = *(ParseEUri(other))
	return
}
