package sip

import (
	"log"
	"bytes"
	"net/url"
	"strings"
)

type parseUriFn func(*lexer) parseUriFn


// Extended URI is an uri with extra parameters and common name
// like: "John Doh" <john.doh@locallab.net>;tag=1234
// usefull for "From", "To" and "Contact"
type EUri struct {
	CommonName string
	U Uri
	Parameters map[string]string
}

type Uri struct {
	Schema string
	User string
	Passwd string
	Host string 
	Port string
	Parameters map[string]string
	Headers map[string]string
}

func ParseUri(s string) *Uri{
	u := new(Uri)
	u.Schema="sip"
	u.Parameters = make(map[string]string)
	u.Headers = make(map[string]string)
	tmp,_ := url.QueryUnescape(strings.TrimSpace(s))
	l := &lexer{
		[]rune(tmp),
		false,
		nil,
		u,
		0,
		0,
	}
	for state:=startParseUri; state!=nil; {
		state = state(l)
	}
	return u
}


func (u *Uri)String() string{
	prepart := false
	b := bytes.NewBufferString(u.Schema)
	b.WriteString(":")
	if len(u.User)>0 {
		prepart = true
		b.WriteString(u.User)
	}
	if len(u.Passwd)>0 {
		prepart = true
		b.WriteString(":"+u.Passwd)
	}
	if len(u.Host)>0 {
		if prepart {
			b.WriteString("@")
		}
		b.WriteString(u.Host)
	}
	if len(u.Port)>0 {
		b.WriteString(":"+u.Port)
	}
	for k,v := range(u.Parameters) {
		b.WriteString(";"+k+"="+v)
	}
	return b.String()
}

// Escaped String
func (u *Uri)EString() string {
	return url.QueryEscape (u.String());
}

func ParseEUri(s string) *EUri{
	u := new(EUri)
	u.U.Schema="sip"
	u.U.Parameters = make(map[string]string)
	u.U.Headers = make(map[string]string)
	u.Parameters = make(map[string]string)
	tmp,_ := url.QueryUnescape(strings.TrimSpace(s))
	l := &lexer{
		[]rune(tmp),
		true,
		u,
		nil,
		0,
		0,
	}
	for state:=startParseEUri; state!=nil; {
		state = state(l)
	}

	return u
}

func (e *EUri)String() string{
	b := bytes.NewBufferString("")
	if e.CommonName != "" {
		b.WriteString("\""+e.CommonName+"\" ")
	}
	b.WriteString("<")
	b.WriteString(e.U.String())
	b.WriteString(">")
	for k,v := range(e.Parameters) {
		b.WriteString(";"+k+"="+v)
	}
	return b.String()
}

func (e *EUri)EString() string {
	return url.QueryEscape(e.String());
}

func (e *EUri)IsEmpty() bool {
	return (e.U.Host=="")
}

// Below this line just implementation details. 
// go on carefully :D

type lexer struct {
	val    []rune
	isEUri bool
	EData *EUri
	Data  *Uri
	start  int
	pos    int
}

func startParseEUri(l *lexer) parseUriFn{
	if l.val[l.pos]=='"'{
		l.pos = l.pos+1
		l.start = l.pos
	}
	l.pos = IndexRune(l.val[l.start:],"\"<")
	if l.pos==-1 {
		return startParseUri
	} else {
		l.pos = l.pos+l.start
	}
	l.EData.CommonName = string(l.val[l.start:l.pos])
	l.start = l.pos+1
	for l.val[l.start] == ' ' {
		l.pos = l.pos+1
		l.start = l.pos
	}
	return startParseUri
}

func startParseUri(l *lexer) parseUriFn{
	log.Println(l.val[l.start]=='<')
	if l.val[l.start]=='<' {
		l.pos = l.pos+1
		l.start = l.pos
	}
	l.pos=IndexRune(l.val[l.start:],":")
	if l.pos==-1 {
		return nameParse
	} else {
		l.pos = l.pos+l.start
	}
	if l.isEUri {
		l.EData.U.Schema = string(l.val[l.start:l.pos])
	} else {
		l.Data.Schema = string(l.val[l.start:l.pos])
	}
	l.start = l.pos+1
	return nameParse
}

func nameParse(l *lexer) parseUriFn{
	l.pos = IndexRune(l.val[l.start:],":@")
	if l.pos==-1 {
		return hostParse
	} else {
		tmp := IndexRune(l.val[l.start:],"@")
		if tmp == -1 {
			return hostParse
		}
		l.pos = l.pos+l.start
	}
	if l.isEUri {
		l.EData.U.User = string(l.val[l.start:l.pos])
	} else {
		l.Data.User = string(l.val[l.start:l.pos])
	}
	l.start = l.pos+1
	if l.pos>=len(l.val) {
		return nil
	} else if l.val[l.pos] == ':' {
		//got passwod
		return passwdParse
	} else if l.val[l.pos] == '@' {
		return hostParse
	} else if l.val[l.pos] == '>' {
		if l.isEUri {
			return parseEParameter
		} else {
			return nil
		}
	}
	return nil
}

func passwdParse(l *lexer) parseUriFn{
	l.pos = IndexRune(l.val[l.start:],"@")
	if l.pos==-1 {
		l.pos = len(l.val)
	} else {
		l.pos = l.pos+l.start
	}
	if l.pos==-1 {
		return nil
	}
	if l.isEUri {
		l.EData.U.Passwd = string(l.val[l.start:l.pos])
	} else {
		l.Data.Passwd = string(l.val[l.start:l.pos])
	}
	l.start = l.pos+1
	return hostParse
}

func hostParse(l *lexer) parseUriFn {
	l.pos = IndexRune(l.val[l.start:],";:>")
	if l.pos==-1 {
		l.pos = len(l.val)
	} else {
		l.pos = l.pos+l.start
	}
	if l.isEUri {
		l.EData.U.Host = string(l.val[l.start:l.pos])
	} else {
		l.Data.Host = string(l.val[l.start:l.pos])
	}
	l.start = l.pos + 1
	if l.pos>=len(l.val) {
		return nil
	} else if l.val[l.pos]==':' {
		return portParse
	} else if l.val[l.pos]==';' {
		return parameterParse
	} else if l.val[l.pos]=='>' {
		if l.isEUri {
			return parseEParameter
		} else {
			return nil
		}		
	}
	return nil
}

func portParse(l *lexer) parseUriFn {
	l.pos = IndexRune(l.val[l.start:],";>")
	if l.pos==-1 {
		l.pos = len(l.val)
	} else {
		l.pos = l.pos+l.start
	}
	if l.isEUri {
		l.EData.U.Port = string(l.val[l.start:l.pos])
	} else {
		l.Data.Port = string(l.val[l.start:l.pos])
	}
	l.start = l.pos + 1
	if l.pos>=len(l.val) {
		return nil
	} else if l.val[l.pos]==';' {
		return parameterParse
	} else if l.val[l.pos]=='>' {
		if l.isEUri {
			return parseEParameter
		}
		return nil
	}
	return nil
}

func parameterParse(l *lexer) parseUriFn {
	l.pos = IndexRune(l.val[l.start:],";?>")
	if l.pos==-1 {
		l.pos = len(l.val)
	} else {
		l.pos = l.pos+l.start
	}
	tmp := string(l.val[l.start:l.pos])
	l.start = l.pos+1
	t := strings.SplitN(tmp, "=", 2)
	if l.isEUri {
		l.EData.U.Parameters[t[0]] = t[1]
	} else {
		l.Data.Parameters[t[0]] = t[1]
	}
	if l.pos>=len(l.val) {
		return nil
	} else if l.val[l.pos]==';' {
		return parameterParse
	} else if l.val[l.pos]=='?' {
		return nil //not implemented
	} else if l.val[l.pos]=='>' {
		if l.isEUri {
			return parseEParameter
		}
		return nil
	}
	return nil
}

func parseEParameter(l *lexer) parseUriFn {
	l.pos = IndexRune(l.val[l.start:],";")
	if l.pos==-1 {
		return nil
	} else {
		l.pos = l.pos+l.start
	}
	l.start = l.pos+1
	return eParameterParse
}

func eParameterParse(l *lexer) parseUriFn {
	l.pos = IndexRune(l.val[l.start:],";")
	if l.pos==-1 {
		l.pos = len(l.val)
	} else {
		l.pos = l.pos+l.start
	}
	tmp := string(l.val[l.start:l.pos])
	if tmp == "" {
		return nil
	}
	l.start = l.pos+1
	t := strings.SplitN(tmp, "=", 2)
	l.EData.Parameters[t[0]] = t[1]
	if l.pos>=len(l.val) {
		return nil
	} else if l.val[l.pos]==';' {
		return parseEParameter
	}
	return nil
}
