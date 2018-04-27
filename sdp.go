//SDP for WebRTC
//https://tools.ietf.org/html/draft-ietf-rtcweb-sdp-09
//we don't parse some lines that webrtc doesn't care about.

package sdp

import (
	"errors"
	"log"
	"strings"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

type SDP struct {
	Version     uint      //v=  (protocol version, must be 0)
	Origin      Origin    //o=  (originator and session identifier)
	SessName    string    //s=  (session name)
	ConnInfo    string    //c=* (connection information -- not required if included in all media)
	Time        TIME      //t=  (time the session is active)
	GroupBundle [3]string //a=group:BUNDLE audio video data (audio video use same port)
	GroupLS     [2]string //a=group:LS audio video  (mid1 mid2)

	// LS
	IceTrickle bool   //a=ice-options:trickle
	Identity   string //a=identity:...
	MsIDSema   string //a=msid-semantic: WMS ...(webrtc media stream id semantic)
	//Media
	Medias []media
}

func (sdp *SDP) Decode(sdpstr string) error {
	if len(sdpstr) == 0 {
		return errors.New("the sdp string is empty!")
	}

	//split sdp into three part: session audio video
	parts := strings.Split(sdpstr, "m=")
	var err string
	for i, part := range parts {
		if i == 0 {
			err += sdp.DecodeGlobal(&part).Error()
			if err != "" {
				log.Println(err)
			}
		} else {
			media := Line("m=" + part)
			if media.Has(MEDIA_VIDEO) || media.Has(MEDIA_AUDIO) {
				sdp.Medias = append(sdp.Medias, &AVMedia{})
			} else if media.Has(MEDIA_DATA_CHAN) {
				sdp.Medias = append(sdp.Medias, &DataChannel{})
			} else {
				return errors.New("unknown media")
			}
			err += sdp.Medias[i-1].Decode(&media)
			if err != "" {
				log.Println(err)
			}

		}
	}
	if err != SDP_ERR_NONE {
		return errors.New(err)
	}
	return nil
}

func (sdp *SDP) Encode() string {
	return sdp.EncodeGlobal() + sdp.EncodeMedia()
}

func (sdp *SDP) DecodeGlobal(s *string) error {
	lines := SplitLines(s)
	var err string
	for i := 0; i < len(*lines); i++ {
		l := (*lines)[i]
		switch l[0] {
		case 'v':
			err += l.Decode("v=%d", &sdp.Version)
		case 'o':
			err += sdp.Origin.Decode(&l)
		case 's':
			err += l.Decode("s=%s", &sdp.SessName)
		case 't':
			err += l.Decode("t=%d %d", &sdp.Time.Start, &sdp.Time.End)
		case 'a':
			if l.Has("group:") {
				err += sdp.decodeGroupLine(&l)
			} else if l.Has("msid-semantic:") {
				err += sdp.decodeMsIDSemaLine(&l)
				//e.g. there is one a=fingerprint: in global, not in the media lines
			} else if l.Has("a=fingerprint:") {
				for i := 0; i < len(sdp.Medias); i++ {
					err += sdp.Medias[i].DecodeDTLSFinger(&l)
				}
				//e.g. there is one a=ice-options:trickle in global, not in the media lines
			} else if l.Has("a=ice-options:") {
				if l.Has("a=ice-options:trickle") {
					sdp.IceTrickle = true
				}
				for i := 0; i < len(sdp.Medias); i++ {
					err += sdp.Medias[i].DecodeICE(&l)
				}
			} else if l.Has("a=identity:") {
				err += l.Decode("a=identity:%s", &sdp.Identity)
			} else {
				err += SdpErrFmt("SDP.DecodeSession", "other session attributes! "+l.Str())
			}
		case 'm':
			err += SdpErrFmt("SDP.DecodeSession", "case m")
		default:
			err += SdpErrFmt("SDP.DecodeSession", "default")
		}
	}
	return errors.New(err)
}

func (sdp *SDP) EncodeGlobal() string {
	l := Line("")
	l.Append("v=%d\r\n", sdp.Version)
	l.Append(sdp.Origin.Encode())
	l.Append("s=%s\r\n", sdp.SessName)
	l.Append("t=%d %d\r\n", sdp.Time.Start, sdp.Time.End)
	l.Append("a=group:BUNDLE %s", sdp.GroupBundle[0])
	l.Append(" %s", sdp.GroupBundle[1])
	l.Append(" %s\r\n", sdp.GroupBundle[2])
	l.Append("a=group:LS %s", sdp.GroupLS[0])
	l.Append(" %s\r\n", sdp.GroupLS[1])
	l.Append("a=msid-semantic: WMS %s\r\n", sdp.MsIDSema)
	l.Append("a=identity:%s\r\n", sdp.Identity)
	l.Append("a=ice-options:trickle\r\n", sdp.IceTrickle)

	return l.Str()
}

func (sdp *SDP) EncodeMedia() string {
	l := Line("")
	for i := 0; i < len(sdp.Medias); i++ {
		sdp.Medias[i].Encode(&l)
	}
	return string(l)
}

//e.g.
//a=msid-semantic: WMS lgsCFqt9kN2fVKw5wg3NKqGdATQoltEwOdMS
func (sdp *SDP) decodeMsIDSemaLine(l *Line) (errStr string) {
	//a=msid-semantic: WMS d4CjVgGfYTSSBBRvhwpj6fDX45NDPwQpQosZ
	if l.Decode("a=msid-semantic:WMS %s", &sdp.MsIDSema) == SDP_ERR_NONE {
		return SDP_ERR_NONE
	} else if l.Has("a=msid-semantic: WMS") {
		sdp.MsIDSema = l.Str()[21:]
		return SDP_ERR_NONE
	}
	return SdpErrFmt("decodeMsIDSemaLine", l.Str())
}

func (sdp *SDP) decodeGroupLine(l *Line) (errStr string) {
	if l.Has("a=group:BUNDLE") {
		//sscanf 1 or 2 var is ok
		l.Decode("a=group:BUNDLE %s %s %s", &sdp.GroupBundle[0], &sdp.GroupBundle[1], &sdp.GroupBundle[2])
		return SDP_ERR_NONE
	}
	if l.Has("a=group:LS") {
		l.Decode("a=group:LS %s %s", &sdp.GroupLS[0], &sdp.GroupLS[1])
		return SDP_ERR_NONE
	}
	return SdpErrFmt("decodeGroupLine", l.Str())
}
