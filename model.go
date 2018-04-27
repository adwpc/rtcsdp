package sdp

import (
	"fmt"
	"net"
	"strings"
)

type Line string

func (l *Line) Has(subStr string) bool {
	if strings.Contains(string(*l), subStr) {
		return true
	}
	return false
}

func (l *Line) Str() string {
	return string(*l)
}

//if param exist append
func (l *Line) Append(format string, a ...interface{}) {
	if format == "" {
		return
	}
	if a == nil {
		*l = Line(l.Str() + format)
		return
	}
	switch a[0].(type) {
	case bool:
		if a[0].(bool) {
			*l = Line(l.Str() + format)
		}
	case string:
		if a[0].(string) != "" {
			*l = Line(l.Str() + fmt.Sprintf(format, a...))
		}
	case uint:
		*l = Line(l.Str() + fmt.Sprintf(format, a...))
	case int:
		if a[0].(int) != -1 {
			*l = Line(l.Str() + fmt.Sprintf(format, a...))
		}
	}
}

func (l *Line) AppendIfHas(flag interface{}, format string, a ...interface{}) {
	switch flag.(type) {
	case bool:
		if flag.(bool) {
			l.Append(format, a...)
		}
	case string:
		if flag.(string) != "" {
			l.Append(format, a...)
		}
	case uint:
		if flag.(uint) != 0 {
			l.Append(format, a...)
		}
	}
}

func (l *Line) Decode(format string, a ...interface{}) (errStr string) {
	_, err := fmt.Sscanf(string(*l), format, a...)
	if err != nil {
		return SdpErrFmt("Line.Decode", l.Str()+"|"+err.Error())
	}
	return SDP_ERR_NONE
}

func (l *Line) DecodeIfHas(flag, format string, a ...interface{}) (errStr string) {
	if l.Has(flag) {
		_, err := fmt.Sscanf(string(*l), format, a...)
		if err != nil {
			return SdpErrFmt("Line.DecodeIfHas", l.Str()+"|"+err.Error())
		}
		return SDP_ERR_NONE
	}
	return SdpErrFmt("Line.DecodeIfHas fail", flag+"|"+l.Str())
}

// func (l *Line) Encode(format string, a ...interface{}) {
// *l += fmt.Sprintf(format+"\r\n", a...)
// log.Println(fmt.Sprintf(format+"\r\n", a...))
// }

//o=<username> <sess-id> <sess-version> <nettype> <addrtype> <unicast-address>
type Origin struct {
	UserName    string
	SessID      string
	SessVersion string
	NetType     string
	AddrType    string
	Address     string
}

func (o *Origin) Decode(l *Line) (errStr string) {
	return l.Decode("o=%s %s %s %s %s %s", &o.UserName, &o.SessID, &o.SessVersion, &o.NetType, &o.AddrType, &o.Address)
}

//e.g.
//o=- 1464328034279435 1464328034279435 IN IP4 192.168.24.158
func (o *Origin) Encode() string {
	return fmt.Sprintf("o=%s %s %s %s %s %s\r\n", o.UserName, o.SessID, o.SessVersion, o.NetType, o.AddrType, o.Address)
}

// type Time struct {
// Start int
// End   int
// }

//e.g.
//a=rtpmap:111 opus/48000/2
//a=rtcp-fb:111 transport-cc
//a=fmtp:111 minptime=10; useinbandfec=1
//a=rtpmap:103 ISAC/16000
//a=rtpmap:104 ISAC/32000
//a=rtpmap:9 G722/8000
//a=rtpmap:0 PCMU/8000
//a=rtpmap:8 PCMA/8000
//a=rtpmap:106 CN/32000
//a=rtpmap:105 CN/16000
//a=rtpmap:13 CN/8000
//a=rtpmap:126 telephone-event/8000
type RTP struct {
	RtpFmtID     string
	RtpFmtVal    string
	RtpFmtParam  string
	RtcpFbParams []string
}

type SSRC struct {
	ID      string
	CName   string
	MsID    string
	MsLabel string
	Label   string
}

//a=sctpmap:50000 webrtc-datachannel 1024
//a=sctpmap:sctpmap-number app [max-message-size] [streams]
type SCTPMAP struct {
	Number     uint
	App        string
	MaxMsgSize uint
	Streams    uint
}

func (s *SCTPMAP) Decode(l *Line) (errStr string) {
	return l.Decode("%d %s %d %d", &s.Number, &s.App, &s.MaxMsgSize, &s.Streams)
}

//c=<nettype> <addrtype> <connection-address>
type Connection struct {
	NetType  string
	AddrType string
	Addr     string
}

//a=rtcp:9 IN IP4 0.0.0.0
type RTCP struct {
	Port    string
	NetType string
	IPType  string
	IP      string
}

type Candidate struct {
	ID         uint   //candidate id
	TransType  string //1 rtp 2 rtcp
	Proto      string //udp tcp
	Priority   uint   //priority, webrtc use bigger one
	IP         net.IP //ip
	Port       uint   //port
	HostType   string //typ host(p2p)    typ srflx(stun)    typ relay(turn)
	TcpType    string //tcptype  e.g. active
	RAddr      net.IP //used by turn or stun
	RPort      uint   //used by turn or stun
	Generation string //generation
	IsEnd      bool   //a=end-of-candidates
}

type DTLS struct {
	Finger     string //fingerprint
	FingerAlgo string //fingerprint algorithm
	Conn       string //connection e.g. a=connection:new
	Setup      string //setup
	TlsID      string //tls-id
}

//e.g.
//a=fingerprint:sha-256 94:42:B8:B8:BA:B4:52:3D:F3:59:4F:6E:E4:82:AC:82:27:1F:95:DA:40:A6:B9:4B:2D:9B:A7:BE:E9:4D:AE:42
//a=setup:actpass
//a=connection:new
//a=tls-id:89J2LRATQ3ULA24G9AHWVR31VJWSLB68

func (d *DTLS) Decode(l *Line) (err string) {
	switch {
	case l.DecodeIfHas("a=fingerprint:", "a=fingerprint:%s %s", &d.FingerAlgo, &d.Finger) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	case l.DecodeIfHas("a=setup:", "a=setup:%s", &d.Setup) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	case l.DecodeIfHas("a=connection:", "a=connection:%s", &d.Conn) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	case l.DecodeIfHas("a=tls-id:", "a=tls-id:%s", &d.TlsID) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	}
	return SdpErrFmt("DTLS.Decode", l.Str())
}

func (d *DTLS) Encode(l *Line) {
	l.Append("a=fingerprint:%s %s\r\n", d.FingerAlgo, d.Finger)
	l.Append("a=setup:%s\r\n", d.Setup)
	l.Append("a=connection:%s\r\n", d.Conn)
	l.Append("a=tls-id:%s\r\n", d.TlsID)
}

type ICE struct {
	Cands    []*Candidate //ICE candidates
	Ufrag    string       //ICE ufrag
	Pwd      string       //ICE pwd
	Options  string       //ICE options e.g. trikle
	Identity string       //https://tools.ietf.org/html/draft-ietf-rtcweb-security-arch-14#section-5.6.4.2
}

//e.g.
//a=ice-ufrag:1eK7W+oSBMWFa8Pe
//a=ice-pwd:V7nbdVnnGGW0F+ZQfiz9841Z
//a=ice-options:trickle
//a=identity:...
func (i *ICE) Decode(l *Line) (errStr string) {
	switch {
	case !l.Has("a=ice") && !l.Has("a=identity"):
		return SdpErrFmt("ICE.Decode", "invalid ice line:"+l.Str())
	case l.DecodeIfHas("ice-ufrag", "a=ice-ufrag:%s", &i.Ufrag) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	case l.DecodeIfHas("ice-pwd", "a=ice-pwd:%s", &i.Pwd) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	case l.DecodeIfHas("ice-options", "a=ice-options:%s", &i.Options) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	case l.DecodeIfHas("ice-identity", "a=ice-identity:%s", &i.Identity) == SDP_ERR_NONE:
		return SDP_ERR_NONE
	}

	return SdpErrFmt("ICE.Decode", l.Str())
}

func (i *ICE) Encode(l *Line) {
	l.Append("a=ice-ufrag:%s\r\n", i.Ufrag)
	l.Append("a=ice-pwd:%s\r\n", i.Pwd)
	l.Append("a=ice-options:%s\r\n", i.Options)
}

type ExtMap struct {
	Key string
	Val string
}

type TIME struct {
	Start uint
	End   uint
}
