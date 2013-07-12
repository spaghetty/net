package sip

type SipEndPoint interface {
	HandleMsg(*SipMsg)
}
