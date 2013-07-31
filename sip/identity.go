package sip

import (
	"errors"
	"strconv"
)

type Identity interface {
	Key() string
	GetEUri() (*EUri, error)
	GetProxy() string
}

type DIdentity struct {
	Name string
	User string
	Password string
	Domain string
	Proxy string
	ProxyPort int
	Registrar string
	RegistrarPort int
}


func NewStdIdentity(uri string, pws string) *DIdentity {
	t := ParseEUri(uri)
	i := new(DIdentity)
	i.Name = t.CommonName
	i.User = t.U.User
	i.Password = pws
	i.Domain = t.U.Host
	if t.U.Port!="" {
		i.Proxy = t.U.Host
		i.ProxyPort,_ = strconv.Atoi(t.U.Port)
	}
	return i
}


func (i *DIdentity) GetProxy() string {
	if i.Proxy!="" {
		if i.ProxyPort!=0 {
			return i.Proxy + ":" + strconv.Itoa(i.ProxyPort)
		}
		return i.Proxy
	}
	return ""
}

func (i *DIdentity) Key() string {
	if v, err:= i.GetEUri(); err!=nil {
		return v.String()
	}
	return ""
}

func (i *DIdentity)GetEUri() (*EUri,error){
	e := new(EUri)
	e.U.Schema = "sip"
	if i.Name!="" {
		e.CommonName = i.Name
	}
	if i.User!="" {
		e.U.User = i.User
	} else {
		return nil, errors.New("identity: User parameter is mandatory")
	}
	if i.Domain!="" {
		e.U.Host = i.Domain
	} else {
		if i.Proxy != "" {
			if i.ProxyPort!=0  {
				e.U.Host = i.Proxy+":"+strconv.Itoa(i.ProxyPort)
			} else {
				e.U.Host = i.Proxy
			}
		} else {
			return nil, errors.New("identity: At least one between Domain and Proxy need to be set")
		}
	}
	return e, nil
}
