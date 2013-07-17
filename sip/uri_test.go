package sip

import (
//	"fmt"
	"testing"
)

func TestUriBase(t *testing.T) {
	s:="sip:merda@cacca.pupu"
	r := ParseUri(s)
	if r.Schema !=  "sip" {
		t.Error("uncorrect schema")
	}
	if r.User != "merda" {
		t.Error("uncorrect User", r.User)
	}
	if r.Host != "cacca.pupu" {
		t.Error("uncorrect Host" , r.Host)
	}
	if s!=r.String() {
		t.Error("Uncorrect Uri Rendering", r.String())
	}
}

func TestUriParameter(t *testing.T) {
	s := "sip:merda@cacca.pupu:4053;tag=1234567;prova=pippo"
	r := ParseUri(s)
	if r.Schema !=  "sip" {
		t.Error("uncorrect schema")
	}
	if r.User != "merda" {
		t.Error("uncorrect User", r.User)
	}
	if r.Host != "cacca.pupu" {
		t.Error("uncorrect Host" , r.Host)
	}
	if r.Port != "4053" {
		t.Error("uncorrect Port",  r.Port)
	}
	if len(r.Parameters)== 2 {
		if (r.Parameters["prova"]!="pippo" ||
			r.Parameters["tag"]!="1234567") {
			t.Error("uncorrect Parameter", r.Parameters)
		}
	} else {
		t.Error("uncorrect Parameter", r.Parameters)
	}
	if r.String()!= s {
		t.Error("uncorrect conversion", r.String())
	}
}

func TestEUriBase(t *testing.T) {
	s:="\"Ciccio Cappuccio\" <sip:merda@cacca.pupu>"
	r := ParseEUri(s)
	if r.CommonName != "Ciccio Cappuccio" {
		t.Error("uncorrect schema", r.CommonName)
	}
	if r.U.Schema !=  "sip" {
		t.Error("uncorrect schema", r.U.Schema)
	}
	if r.U.User != "merda" {
		t.Error("uncorrect User", r.U.User)
	}
	if r.U.Host != "cacca.pupu" {
		t.Error("uncorrect Host" , r.U.Host)
	}
	if s!=r.String() {
		t.Error("Uncorrect Uri Rendering", r.String())
	}
}

func TestEUriContact(t *testing.T) {
	s:="<sip:cacca.pupu:8090>"
	r := ParseEUri(s)
	if r.U.Schema !=  "sip" {
		t.Error("uncorrect schema", r.U.Schema)
	}
	if r.U.Host != "cacca.pupu" {
		t.Error("uncorrect Host" , r.U.Host)
	}
	if r.U.Port != "8090" {
		t.Error("uncorrect Port" , r.U.Port)
	}
	if s!=r.String() {
		t.Error("Uncorrect Uri Rendering", r.String())
	}
}

func TestEUriParameter(t *testing.T) {
	s := "\"Pippo Pluto\" <sip:merda@cacca.pupu:4053;tag=1234567;prova=pippo>;tag=1234"
	r := ParseEUri(s)
	if r.U.Schema !=  "sip" {
		t.Error("uncorrect schema", r.U.Schema)
	}
	if r.U.User != "merda" {
		t.Error("uncorrect User", r.U.User)
	}
	if r.U.Host != "cacca.pupu" {
		t.Error("uncorrect Host" , r.U.Host)
	}
	if r.U.Port != "4053" {
		t.Error("uncorrect Port",  r.U.Port)
	}
	if len(r.U.Parameters)== 2 {
		if (r.U.Parameters["prova"]!="pippo" ||
			r.U.Parameters["tag"]!="1234567") {
			t.Error("uncorrect Parameter", r.U.Parameters)
		}
	} else {
		t.Error("uncorrect Parameter", r.U.Parameters)
	}
	if len(r.Parameters)==1 {
		if r.Parameters["tag"] != "1234" {
			t.Error("uncorrect E Parameter", r.Parameters)
		}
	} else {
		t.Error("uncorrect E Parameter", r.Parameters)
	}
	if r.String()!= s {
		t.Error("uncorrect conversion", r.String())
	}
}
