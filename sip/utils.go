package sip

import (
	"strings"
)

const TagCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"

type MethodType int

const  (
	INVITE MethodType = iota
	NOTIFY MethodType = iota
)

func getDialog(s SipMsg, t string) (d Dialog) {
	switch s.(type) {
	case *Request:
		return getdialog((s.(*Request)).Header,t)
	case *Response:
		return getdialog((s.(*Response)).Header,t)
	}
	return 
}

func getdialog(h Header, t string) (d Dialog) {
	d.CallId = h["Call-Id"][0]
	tmp := ParseEUri(h["To"][0])
	if v,ok := tmp.Parameters["tag"]; !ok || v==t {
		d.Me = *tmp
		d.Other = *ParseEUri(h["From"][0])
	} else {
		d.Other = *tmp
		d.Me = *ParseEUri(h["From"][0])
	}
	return
}


func IndexRune(r []rune, s string) int {
	for i,v := range(r) {
		if strings.ContainsRune(s,v) {
			return i
		}
	}
	return -1
}
