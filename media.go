package sdp

import (
	"log"
	"net"
	"strings"
)

type media interface {
	//decode
	Decode(l *Line) string
	DecodeDTLSFinger(l *Line) string
	DecodeICE(l *Line) string

	//encode
	Encode(l *Line)
}

type MediaBase struct {
	//m=...
	Type   string //media type, audio video data ...
	Port   string //transport port
	Proto  string //transport protocol
	Format string //for DataChannel , audio video use rtpmap

	//c=...
	ConnData Connection //connection
	//a=rtcp:1 IN IP4 0.0.0.0
	Rtcp      RTCP //rtcp
	RtcpRsize bool
	//a=sendrecv
	Mode string //sendonly recvonly sendrecv
	//a=rtcp-mux
	RtcpMux     bool   //rtp and rtcp can use the same port
	RtcpMuxOnly bool   //rtp and rtcp can use the same port only
	BundleOnly  bool   //a=bundle-only
	Ice         ICE    //ICE
	Dtls        DTLS   //DTLS
	Mid         string //mid

	//如果ssrc>1才会有这行  =1不会有这行
	//a=ssrc-group:FID 2231627014 632943048
	// SSRCFIDS []string         //ssrc group fid array
	SSRCS map[string]*SSRC //Synchronisation Source

	MsID string //firefox webrtc media stream id
}

func (m *MediaBase) DecodeDTLSFinger(l *Line) (errStr string) {
	return l.Decode("a=fingerprint:%s %s", &m.Dtls.FingerAlgo, &m.Dtls.Finger)
}

func (m *MediaBase) DecodeICE(l *Line) (errStr string) {
	return m.Ice.Decode(l)
}

type AVMedia struct {
	MediaBase
	ExtMaps  []ExtMap        //extmap
	RTPMap   map[string]*RTP //codec parameters
	Maxptime string          //audio codec maxptime
}

func (m *AVMedia) Encode(l *Line) {
	l.Append("%s %s %s", m.Type, m.Port, m.Proto)
	for k, _ := range m.RTPMap {
		l.Append(" %s", k)
	}
	l.Append("\r\n")
	l.Append("c=%s %s %s\r\n", m.ConnData.NetType, m.ConnData.AddrType, m.ConnData.Addr)
	l.Append("a=mid:%s\r\n", m.Mid)
	l.Append("a=msid:%s\r\n", m.MsID)
	l.Append("a=%s\r\n", m.Mode)
	l.Append("a=rtcp:%s %s %s %s\r\n", m.Rtcp.Port, m.Rtcp.NetType, m.Rtcp.IPType, m.Rtcp.IP)
	l.Append("a=rtcp-rsize\r\n", m.RtcpRsize)
	l.Append("a=rtcp-mux\r\n", m.RtcpMux)
	l.Append("a=rtcp-mux-only\r\n", m.RtcpMuxOnly)
	l.Append("a=bundle-only\r\n", m.BundleOnly)
	l.Append("a=maxptime:%s\r\n", m.Maxptime)

	m.Ice.Encode(l)
	m.Dtls.Encode(l)
	for i := 0; i < len(m.ExtMaps); i++ {
		l.Append("a=extmap:%s %s\r\n", m.ExtMaps[i].Key, m.ExtMaps[i].Val)
	}
	m.encodeRTPLine(l)
	m.encodeSSRCLine(l)
	m.encodeCandLine(l)

}

func (m *MediaBase) encodeCandLine(l *Line) {
	for i := 0; i < len(m.Ice.Cands); i++ {

		cand := m.Ice.Cands[i]
		if cand.IsEnd {
			l.Append("a=end-of-candidates\r\n")
			return
		}
		switch cand.HostType {
		case CAND_HOST:
			if strings.ToLower(cand.Proto) == CAND_UDP {
				//a=candidate:1467250027 1 udp 2122260223 192.168.0.196 46243 typ host generation 0
				if cand.Generation != "" {
					l.Append("a=candidate:%d %s udp %d %s %d typ %s generation %s\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.HostType, cand.Generation)
				} else {
					l.Append("a=candidate:%d %s udp %d %s %d typ %s\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.HostType)
				}
			} else if strings.ToLower(cand.Proto) == CAND_TCP {
				if cand.Generation != "" {
					l.Append("a=candidate:%d %s tcp %d %s %d typ %s tcptype %s generation %s\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.HostType, cand.TcpType, cand.Generation)
				} else {
					l.Append("a=candidate:%d %s tcp %d %s %d typ %s tcptype %s\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.HostType, cand.TcpType)
				}
			}
		case CAND_SRFLX:
			//a=candidate:1853887674 1 udp 1518280447 47.61.61.61 36768 typ srflx raddr 192.168.0.196 rport 36768 generation 0
			//a=candidate:1 2 UDP 1685987071 203.0.113.141 60065 typ srflx raddr 192.0.2.4 rport 61667
			if cand.Generation != "" {
				l.Append("a=candidate:%d %s udp %d %s %d typ srflx raddr %s rport %d generation %s\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.RAddr, cand.RPort, cand.Generation)
			} else {
				//a=candidate:1 2 UDP 1685987071 203.0.113.141 60065 typ srflx raddr 192.0.2.4 rport 61667
				l.Append("a=candidate:%d %s udp %d %s %d typ srflx raddr %s rport %d\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.RAddr, cand.RPort)
			}
		case CAND_RELAY:
			//a=candidate:750991856 2 udp 25108222 237.30.30.30 51472 typ relay raddr 47.61.61.61 rport 54763 generation 0
			if cand.Generation != "" {
				l.Append("a=candidate:%d %s udp %d %s %d typ relay raddr %s rport %d generation %s\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.RAddr, cand.RPort, cand.Generation)
			} else {
				l.Append("a=candidate:%d %s udp %d %s %d typ relay raddr %s rport %d\r\n", cand.ID, cand.TransType, cand.Priority, cand.IP.String(), cand.Port, cand.RAddr, cand.RPort)
			}
		default:
			log.Println("unknown host type!!!!!!!!!!!")
		}
	}

}

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
func (m *AVMedia) newRtpIfNotExist(key string) {
	if _, ok := m.RTPMap[key]; !ok {
		m.RTPMap[key] = new(RTP)
	}
}

func (m *AVMedia) decodeRTPLine(l *Line) (errStr string) {
	var key, val string

	if l.DecodeIfHas("a=rtpmap:", "a=rtpmap:%s %s", &key, &val) == SDP_ERR_NONE {
		m.newRtpIfNotExist(key)
		m.RTPMap[key].RtpFmtID, m.RTPMap[key].RtpFmtVal = key, val
		return SDP_ERR_NONE
	} else if l.DecodeIfHas("a=rtcp-fb:", "a=rtcp-fb:%s %s", &key, &val) == SDP_ERR_NONE {
		m.newRtpIfNotExist(key)
		if m.Type == MEDIA_AUDIO {
			m.RTPMap[key].RtcpFbParams = append(m.RTPMap[key].RtcpFbParams, val)
			return SDP_ERR_NONE
		} else if m.Type == MEDIA_VIDEO {
			items := strings.Split(strings.Split(l.Str(), "a=rtcp-fb:")[1], " ")
			var str string
			for i := 1; i < len(items); i++ {
				str = str + items[i]
				if i != len(items)-1 {
					str = str + " "
				}
			}

			m.RTPMap[items[0]].RtcpFbParams = append(m.RTPMap[items[0]].RtcpFbParams, str)
			return SDP_ERR_NONE
		} else {
			return SdpErrFmt("other media type", l.Str())
		}
	} else if l.Has("a=fmtp:") {
		if m.Type == MEDIA_AUDIO {
			items := strings.Split(strings.SplitN(l.Str(), "a=fmtp:", 2)[1], " ")
			for i := 1; i < len(items); i++ {
				m.RTPMap[items[0]].RtpFmtParam = m.RTPMap[items[0]].RtpFmtParam + items[i]
				if i < len(items)-1 {
					m.RTPMap[items[0]].RtpFmtParam = m.RTPMap[items[0]].RtpFmtParam + " "
				}
			}
		} else if m.Type == MEDIA_VIDEO {
			items := strings.Split(strings.SplitN(l.Str(), "a=fmtp:", 2)[1], " ")
			for i := 1; i < len(items); i++ {
				if _, ok := m.RTPMap[items[0]]; !ok {
					r := new(RTP)
					m.RTPMap[items[0]] = r
				}
				m.RTPMap[items[0]].RtpFmtParam = m.RTPMap[items[0]].RtpFmtParam + items[i]
				if i < len(items)-1 {
					m.RTPMap[items[0]].RtpFmtParam = m.RTPMap[items[0]].RtpFmtParam + " "
				}
			}
		} else {
			return SdpErrFmt("AVMedia.decodeRTPLine:other media type", l.Str())
		}
		return SDP_ERR_NONE
	}

	return SdpErrFmt("AVMedia.decodeRTPLine", l.Str())
}

func (m *AVMedia) encodeRTPLine(l *Line) {
	for k, v := range m.RTPMap {
		l.Append("a=rtpmap:%s %s\r\n", k, v.RtpFmtVal)
		for i := 0; i < len(v.RtcpFbParams); i++ {
			l.Append("a=rtcp-fb:%s %s\r\n", k, v.RtcpFbParams[i])
		}

		if v.RtpFmtParam != "" {
			l.Append("a=fmtp:%s %s\r\n", k, v.RtpFmtParam)
		}
	}
}

//e.g.
//a=ssrc-group:FID 65275558 190423554
//a=ssrc:65275558 cname:+0VBE2jiQiCaYE/E
//a=ssrc:65275558 msid:d4CjVgGfYTSSBBRvhwpj6fDX45NDPwQpQosZ fdfe23cd-ad58-4029-8011-126b1c245296
//a=ssrc:65275558 mslabel:d4CjVgGfYTSSBBRvhwpj6fDX45NDPwQpQosZ
//a=ssrc:65275558 label:fdfe23cd-ad58-4029-8011-126b1c245296
//a=ssrc:190423554 cname:+0VBE2jiQiCaYE/E
//a=ssrc:190423554 msid:d4CjVgGfYTSSBBRvhwpj6fDX45NDPwQpQosZ fdfe23cd-ad58-4029-8011-126b1c245296
//a=ssrc:190423554 mslabel:d4CjVgGfYTSSBBRvhwpj6fDX45NDPwQpQosZ
//a=ssrc:190423554 label:fdfe23cd-ad58-4029-8011-126b1c245296
func (m *AVMedia) decodeSSRCLine(l *Line) (errStr string) {
	if l.Has("a=ssrc:") {
		if m.SSRCS == nil {
			m.SSRCS = make(map[string]*SSRC)
		}
		items := strings.Split(strings.Split(l.Str(), "a=ssrc:")[1], " ")
		if strings.Contains(items[1], "cname") {
			ssrc := new(SSRC)
			ssrc.ID = items[0]
			ssrc.CName = strings.Split(items[1], "cname:")[1]
			m.SSRCS[items[0]] = ssrc
		} else if strings.Contains(items[1], "msid") {
			m.SSRCS[items[0]].MsID = strings.Split(items[1], "msid:")[1] + " " + items[2]
		} else if strings.Contains(items[1], "mslabel") {
			m.SSRCS[items[0]].MsLabel = strings.Split(items[1], "mslabel:")[1]
		} else if strings.Contains(items[1], "label") {
			m.SSRCS[items[0]].Label = strings.Split(items[1], "label:")[1]
		}
		return SDP_ERR_NONE
	} else if l.Has("a=ssrc-group:") {
		return SDP_ERR_NONE
	}
	return SdpErrFmt("AVMedia.decodeSSRCLine", l.Str())
}

func (m *AVMedia) encodeSSRCLine(l *Line) {
	if len(m.SSRCS) > 1 {
		l.Append("a=ssrc-group:FID")
		for k, _ := range m.SSRCS {
			l.Append(" %s", k)
		}
		l.Append("\r\n")
	}
	for _, v := range m.SSRCS {
		l.Append("a=ssrc:%s cname:%s\r\n", v.ID, v.CName)
		l.Append("a=ssrc:%s msid:%s\r\n", v.ID, v.MsID)
		l.Append("a=ssrc:%s mslabel:%s\r\n", v.ID, v.MsLabel)
		l.Append("a=ssrc:%s label:%s\r\n", v.ID, v.Label)
	}
}

//e.g.
//a=candidate:1467250027 1 udp 2122260223 192.168.0.196 46243 typ host generation 0
//a=candidate:1467250027 2 udp 2122260222 192.168.0.196 56280 typ host generation 0
//a=candidate:435653019 1 tcp 1845501695 192.168.0.196 0 typ host tcptype active generation 0
//a=candidate:435653019 2 tcp 1845501695 192.168.0.196 0 typ host tcptype active generation 0
//a=candidate:1853887674 1 udp 1518280447 47.61.61.61 36768 typ srflx raddr 192.168.0.196 rport 36768 generation 0
//a=candidate:1853887674 2 udp 1518280447 47.61.61.61 36768 typ srflx raddr 192.168.0.196 rport 36768 generation 0
//a=candidate:750991856 2 udp 25108222 237.30.30.30 51472 typ relay raddr 47.61.61.61 rport 54763 generation 0
//a=candidate:750991856 1 udp 25108223 237.30.30.30 58779 typ relay raddr 47.61.61.61 rport 54761 generation 0

//short
//a=candidate:0 1 UDP 2122194687 192.0.2.4 61665 typ host

func (m *MediaBase) decodeCandLine(l *Line) (errStr string) {
	if !l.Has("a=candidate") && !l.Has("end-of-candidates") {
		return "invalid candidate string:" + l.Str()
	}

	ice := &m.Ice
	if l.Has("end-of-candidates") {
		ice.Cands = append(ice.Cands, &Candidate{IsEnd: true})
		return SDP_ERR_NONE
	}
	str := (*l).Str()[12:]
	items := strings.Split(str, " ")

	if len(items) < 7 {
		return "invalid candidate string"
	}

	cand := Candidate{
		ID:        UInt(items[0]),
		TransType: items[1],
		Proto:     items[2],
		Priority:  UInt(items[3]),
		IP:        net.ParseIP(items[4]),
		Port:      UInt(items[5]),
		HostType:  items[7]}
	switch cand.HostType {
	case CAND_HOST:
		if cand.Proto == CAND_UDP {
			if l.Has("generation") && len(items) >= 9 {
				cand.Generation = items[9]
			}
		} else if cand.Proto == CAND_TCP {
			if l.Has("tcptype") && len(items) >= 9 {
				cand.TcpType = items[9]
			}
			if l.Has("generation") && len(items) >= 11 {
				cand.Generation = items[11]
			}
		}
	case CAND_SRFLX:
		cand.RAddr = net.ParseIP(items[9])
		cand.RPort = UInt(items[11])
		if l.Has("generation") && len(items) >= 13 {
			cand.Generation = items[13]
		}
	case CAND_RELAY:
		cand.RAddr = net.ParseIP(items[9])
		cand.RPort = UInt(items[11])
		if l.Has("generation") && len(items) >= 13 {
			cand.Generation = items[13]
		}
	default:
		log.Println("unknown host type!!!")
		return "unknown host type!!!"
	}

	ice.Cands = append(ice.Cands, &cand)
	return SDP_ERR_NONE
}

func (m *AVMedia) Decode(l *Line) (errStr string) {
	str := l.Str()
	lines := SplitLines(&str)
	var err string
	for i := 0; i < len(*lines); i++ {
		line := (*lines)[i]
		switch {
		case line.Has(MEDIA_VIDEO) || line.Has(MEDIA_AUDIO):
			items := strings.Split(string(line), " ")
			if len(items) < 4 {
				return SdpErrFmt("AVMedia.Decode error", line.Str())
			}
			m.Type, m.Port, m.Proto = items[0], items[1], items[2]
			if m.RTPMap == nil {
				m.RTPMap = make(map[string]*RTP)
			}
			for i := 3; i < len(items); i++ {
				m.RTPMap[items[i]] = new(RTP)
			}
		case line.Has("c=") && !line.Has("useinbandfec=1"):
			err += line.Decode("c=%s %s %s", &m.ConnData.NetType, &m.ConnData.AddrType, &m.ConnData.Addr)
		case line.Has("a=rtcp:"):
			err += line.Decode("a=rtcp:%s %s %s %s", &m.Rtcp.Port, &m.Rtcp.NetType, &m.Rtcp.IPType, &m.Rtcp.IP)
		case line.Has("a=rtcp-rsize"):
			m.RtcpRsize = true
		case line.Has("a=ice"):
			err += m.Ice.Decode(&line)
		case line.Has("a=fingerprint:") || line.Has("a=setup:") || line.Has("a=connection:") || line.Has("a=tls-id"):
			err += m.Dtls.Decode(&line)
		case line.Has("a=mid:"):
			err += line.Decode("a=mid:%s", &m.Mid)
		case line.Has("a=msid:"):
			//a=msid:ma ta       //sscanf not suport blank
			// err += line.Decode("a=msid:%s", &m.MsID)
			m.MsID = line.Str()[7:]
		case line.Has("a=extmap:"):
			var a ExtMap
			err += line.Decode("a=extmap:%s %s", &a.Key, &a.Val)
			m.ExtMaps = append(m.ExtMaps, a)
		case line.Has("a=sendonly"):
			m.Mode = "sendonly"
		case line.Has("a=recvonly"):
			m.Mode = "recvonly"
		case line.Has("a=sendrecv"):
			m.Mode = "sendrecv"
		case line.Has("a=rtcp-mux-only"):
			m.RtcpMuxOnly = true
		case line.Has("a=rtcp-mux"):
			m.RtcpMux = true
		case line.Has("a=bundle-only"):
			m.BundleOnly = true
		case line.Has("a=rtpmap") || line.Has("a=rtcp-fb") || line.Has("a=fmtp"):
			err += m.decodeRTPLine(&line)
		case line.DecodeIfHas("a=maxptime", "a=maxptime:%s", &m.Maxptime) == SDP_ERR_NONE:
			err += SDP_ERR_NONE
		case line.Has("a=ssrc:") || line.Has("a=ssrc-group:FID"):
			err += m.decodeSSRCLine(&line)
		case line.Has("a=candidate:"):
			err += m.decodeCandLine(&line)
		case line.Has("a=end-of-candidates"):
			err += m.decodeCandLine(&line)

		}
	}
	return err
}

type DataChannel struct {
	MediaBase
	SctpMap SCTPMAP

	//a=sctp-port:5000
	//a=max-message-size:100000
	SctpPort   int
	MaxMsgSize int
}

// func (m *DataChannel) Decode(l *Line) (errStr string) {
// str := l.Str()
// lines := SplitLines(&str)
// var err string
// for i := 0; i < len(*lines); i++ {
// line := (*lines)[i]
// switch {
// case line.Has(MEDIA_DATA_CHAN):
// //m=application 20000 UDP/DTLS/SCTP webrtc-datachannel
// err += line.Decode("%s %s %s %s", &m.Type, &m.Port, &m.Proto, &m.Format)
// case line.Has("c=") && !line.Has("useinbandfec=1"):
// err += line.Decode("c=%s %s %s", &m.ConnData.NetType, &m.ConnData.AddrType, &m.ConnData.Addr)
// case line.Has("a=mid:"):
// err += line.Decode("a=mid:%s", &m.Mid)
// case line.Has("a=sctp-port:"):
// err += line.Decode("a=sctp-port:%d", &m.SctpPort)
// case line.Has("a=max-message-size:"):
// err += line.Decode("a=max-message-size:%d", &m.MaxMsgSize)
// case line.Has("a=fingerprint:") || line.Has("a=setup:") || line.Has("a=connection:") || line.Has("a=tls-id"):
// err += m.Dtls.Decode(&line)
// case line.Has("a=sendonly"):
// m.Mode = "sendonly"
// case line.Has("a=recvonly"):
// m.Mode = "recvonly"
// case line.Has("a=sendrecv"):
// m.Mode = "sendrecv"
// case line.Has("a=ice"):
// err += m.Ice.Decode(&line)
// case line.Has("a=candidate:") || line.Has("a=end-of-candidates"):
// err += m.decodeCandLine(&line)
// case line.Has("a=sctpmap:"):
// err += m.decodeSctpLine(&line)

// }
// }
// return err
// }

func (m *DataChannel) Decode(l *Line) (errStr string) {
	str := l.Str()
	lines := SplitLines(&str)
	var err string
	for i := 0; i < len(*lines); i++ {
		line := (*lines)[i]
		switch {
		case line.DecodeIfHas(MEDIA_DATA_CHAN, "%s %s %s %s", &m.Type, &m.Port, &m.Proto, &m.Format) == SDP_ERR_NONE:
		case line.Has("c=") && !line.Has("useinbandfec=1"):
			err += line.Decode("c=%s %s %s", &m.ConnData.NetType, &m.ConnData.AddrType, &m.ConnData.Addr)
		case line.DecodeIfHas("a=mid:", "a=mid:%s", &m.Mid) == SDP_ERR_NONE:
		case line.DecodeIfHas("a=sctp-port:", "a=sctp-port:%d", &m.SctpPort) == SDP_ERR_NONE:
		case line.DecodeIfHas("a=max-message-size:", "a=max-message-size:%d", &m.MaxMsgSize) == SDP_ERR_NONE:
		case line.Has("a=fingerprint:") || line.Has("a=setup:") || line.Has("a=connection:") || line.Has("a=tls-id"):
			err += m.Dtls.Decode(&line)
		case line.DecodeIfHas("a=sendonly", "a=%s", &m.Mode) == SDP_ERR_NONE:
		case line.DecodeIfHas("a=recvonly", "a=%s", &m.Mode) == SDP_ERR_NONE:
		case line.DecodeIfHas("a=sendrecv", "a=%s", &m.Mode) == SDP_ERR_NONE:
		case line.Has("a=ice"):
			err += m.Ice.Decode(&line)
		case line.Has("a=candidate:") || line.Has("a=end-of-candidates"):
			err += m.decodeCandLine(&line)
		case line.Has("a=sctpmap:"):
			err += m.decodeSctpLine(&line)

		}
	}
	return err
}

func (m *DataChannel) Encode(l *Line) {
	l.Append("%s %s %s %s\r\n", m.Type, m.Port, m.Proto, m.Format)
	l.Append("c=%s %s %s\r\n", m.ConnData.NetType, m.ConnData.AddrType, m.ConnData.Addr)
	l.Append("a=mid:%s\r\n", m.Mid)
	if m.SctpPort != 0 {
		l.Append("a=sctp-port:%d\r\n", m.SctpPort)
	}
	if m.MaxMsgSize != 0 {
		l.Append("a=max-message-size:%d\r\n", m.MaxMsgSize)
	}
	m.Dtls.Encode(l)
	l.Append("a=%s\r\n", m.Mode)
	l.Append("a=sctpmap:%d %s %d\r\n", m.SctpMap.Number, m.SctpMap.App, m.SctpMap.MaxMsgSize)

	m.Ice.Encode(l)
	m.encodeCandLine(l)

}

func (m *DataChannel) decodeSctpLine(l *Line) (errStr string) {
	// 2018/04/27 15:14:20 media.go:435: ---------- a=sctpmap:5000 webrtc-datachannel 1024
	if err := l.DecodeIfHas("a=sctpmap:", "a=sctpmap:%d %s %d", &m.SctpMap.Number, &m.SctpMap.App, &m.SctpMap.MaxMsgSize); err != SDP_ERR_NONE {
		log.Println(err)
		return err
	}
	return SDP_ERR_NONE
}
