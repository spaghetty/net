package sip

type Stack struct {
	cnx    *conn
	callid string
	Handler SipHandler
}

func NewStack(c *conn, id string , h SipHandler) *Stack {
	s := &Stack{c,
		id,
		h}
	return s
}