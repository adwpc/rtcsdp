[![Build Status](https://travis-ci.org/adwpc/rtcsdp.svg?branch=master)](https://travis-ci.org/adwpc/rtcsdp)

# rtcsdp for golang
[draft-nandakumar-rtcweb-sdp](https://tools.ietf.org/html/draft-ietf-rtcweb-sdp-09)


### Examples

```
package main

import (
	"log"

	"github.com/adwpc/rtcsdp"
)

const chromeOffer = `
v=0
o=- 2670773066989454674 2 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS X0EclBo2L6ejdTpXYKIgSFxeEeuWSgrgkJxf
m=audio 9 UDP/TLS/RTP/SAVPF 111 103 104 9 0 8 106 105 13 110 112 113 126
c=IN IP4 0.0.0.0
a=rtcp:9 IN IP4 0.0.0.0
a=ice-ufrag:MsTv
a=ice-pwd:ldzjCz3mvZHU7bdmQdvG9E7R
a=ice-options:trickle
a=fingerprint:sha-256 A1:3E:57:FD:06:C7:C2:75:7D:39:C7:EC:2C:94:EB:CD:D3:94:AD:6B:73:F7:68:31:63:B5:4F:81:43:CC:B7:82
a=setup:actpass
a=mid:audio
a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level
a=sendrecv
a=rtcp-mux
a=rtpmap:111 opus/48000/2
a=rtcp-fb:111 transport-cc
a=fmtp:111 minptime=10;useinbandfec=1
a=rtpmap:103 ISAC/16000
a=rtpmap:104 ISAC/32000
a=rtpmap:9 G722/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:106 CN/32000
a=rtpmap:105 CN/16000
a=rtpmap:13 CN/8000
a=rtpmap:110 telephone-event/48000
a=rtpmap:112 telephone-event/32000
a=rtpmap:113 telephone-event/16000
a=rtpmap:126 telephone-event/8000
a=ssrc:2674873303 cname:k40qiudEPhB/+NLQ
a=ssrc:2674873303 msid:X0EclBo2L6ejdTpXYKIgSFxeEeuWSgrgkJxf 96deaecd-ec83-4324-9eed-51641c51166d
a=ssrc:2674873303 mslabel:X0EclBo2L6ejdTpXYKIgSFxeEeuWSgrgkJxf
a=ssrc:2674873303 label:96deaecd-ec83-4324-9eed-51641c51166d
m=video 9 UDP/TLS/RTP/SAVPF 96 97 98 99 100 101 102 123 127 122 125 107 108 109 124
c=IN IP4 0.0.0.0
a=rtcp:9 IN IP4 0.0.0.0
a=ice-ufrag:MsTv
a=ice-pwd:ldzjCz3mvZHU7bdmQdvG9E7R
a=ice-options:trickle
a=fingerprint:sha-256 A1:3E:57:FD:06:C7:C2:75:7D:39:C7:EC:2C:94:EB:CD:D3:94:AD:6B:73:F7:68:31:63:B5:4F:81:43:CC:B7:82
a=setup:actpass
a=mid:video
a=extmap:2 urn:ietf:params:rtp-hdrext:toffset
a=extmap:3 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time
a=extmap:4 urn:3gpp:video-orientation
a=extmap:5 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01
a=extmap:6 http://www.webrtc.org/experiments/rtp-hdrext/playout-delay
a=extmap:7 http://www.webrtc.org/experiments/rtp-hdrext/video-content-type
a=extmap:8 http://www.webrtc.org/experiments/rtp-hdrext/video-timing
a=sendrecv
a=rtcp-mux
a=rtcp-rsize
a=rtpmap:96 VP8/90000
a=rtcp-fb:96 goog-remb
a=rtcp-fb:96 transport-cc
a=rtcp-fb:96 ccm fir
a=rtcp-fb:96 nack
a=rtcp-fb:96 nack pli
a=rtpmap:97 rtx/90000
a=fmtp:97 apt=96
a=rtpmap:98 VP9/90000
a=rtcp-fb:98 goog-remb
a=rtcp-fb:98 transport-cc
a=rtcp-fb:98 ccm fir
a=rtcp-fb:98 nack
a=rtcp-fb:98 nack pli
a=rtpmap:99 rtx/90000
a=fmtp:99 apt=98
a=rtpmap:100 H264/90000
a=rtcp-fb:100 goog-remb
a=rtcp-fb:100 transport-cc
a=rtcp-fb:100 ccm fir
a=rtcp-fb:100 nack
a=rtcp-fb:100 nack pli
a=fmtp:100 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f
a=rtpmap:101 rtx/90000
a=fmtp:101 apt=100
a=rtpmap:102 H264/90000
a=rtcp-fb:102 goog-remb
a=rtcp-fb:102 transport-cc
a=rtcp-fb:102 ccm fir
a=rtcp-fb:102 nack
a=rtcp-fb:102 nack pli
a=fmtp:102 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f
a=rtpmap:123 rtx/90000
a=fmtp:123 apt=102
a=rtpmap:127 H264/90000
a=rtcp-fb:127 goog-remb
a=rtcp-fb:127 transport-cc
a=rtcp-fb:127 ccm fir
a=rtcp-fb:127 nack
a=rtcp-fb:127 nack pli
a=fmtp:127 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=4d0032
a=rtpmap:122 rtx/90000
a=fmtp:122 apt=127
a=rtpmap:125 H264/90000
a=rtcp-fb:125 goog-remb
a=rtcp-fb:125 transport-cc
a=rtcp-fb:125 ccm fir
a=rtcp-fb:125 nack
a=rtcp-fb:125 nack pli
a=fmtp:125 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=640032
a=rtpmap:107 rtx/90000
a=fmtp:107 apt=125
a=rtpmap:108 red/90000
a=rtpmap:109 rtx/90000
a=fmtp:109 apt=108
a=rtpmap:124 ulpfec/90000
a=ssrc-group:FID 3122697176 552506581
a=ssrc:3122697176 cname:k40qiudEPhB/+NLQ
a=ssrc:3122697176 msid:X0EclBo2L6ejdTpXYKIgSFxeEeuWSgrgkJxf 8b27d69c-993d-4f36-a529-dbcd15fc7e0e
a=ssrc:3122697176 mslabel:X0EclBo2L6ejdTpXYKIgSFxeEeuWSgrgkJxf
a=ssrc:3122697176 label:8b27d69c-993d-4f36-a529-dbcd15fc7e0e
a=ssrc:552506581 cname:k40qiudEPhB/+NLQ
a=ssrc:552506581 msid:X0EclBo2L6ejdTpXYKIgSFxeEeuWSgrgkJxf 8b27d69c-993d-4f36-a529-dbcd15fc7e0e
a=ssrc:552506581 mslabel:X0EclBo2L6ejdTpXYKIgSFxeEeuWSgrgkJxf
a=ssrc:552506581 label:8b27d69c-993d-4f36-a529-dbcd15fc7e0e
m=application 9 DTLS/SCTP 5000
c=IN IP4 0.0.0.0
a=ice-ufrag:MsTv
a=ice-pwd:ldzjCz3mvZHU7bdmQdvG9E7R
a=ice-options:trickle
a=fingerprint:sha-256 A1:3E:57:FD:06:C7:C2:75:7D:39:C7:EC:2C:94:EB:CD:D3:94:AD:6B:73:F7:68:31:63:B5:4F:81:43:CC:B7:82
a=setup:actpass
a=mid:data
a=sctpmap:5000 webrtc-datachannel 1024
`

func main() {
	log.Println("==================decode======================")

	if sdp, err := sdp.NewSDP(chromeOffer); err == nil {
		log.Println("==================encode======================")
		log.Println(sdp.Encode())
	} else {
		log.Println(err.Error())
	}
}

```

### tested sdp
- [x] chrome_avd.sdp
	- audio
	- video
	- data
	
- [x] firefox_av.sdp
	- audio
	- video
- [x] janus_av.sdp
	- audio
	- video

	




