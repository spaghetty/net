package sip

import (
	"strings"
	"net/textproto"
)

func isSipStart(line string) bool {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return false
	}
	s2 += s1 + 1
	return (line[:s1]=="SIP/2.0" || line[s2+1:]=="SIP/2.0")
}

type SipMsg struct {
	StartLine string
	Headers   textproto.MIMEHeader
	Body      string
}

func (m *SipMsg)IsRequest() bool {
	return !strings.HasPrefix(m.StartLine, "SIP/")
}

