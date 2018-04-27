package sdp

const (
	SDP_ERR_NONE = ""
)

//candidate
const (
	CAND_UDP = "udp"
	CAND_TCP = "tcp"
)
const (
	CAND_RTP  = "rtp"
	CAND_RTCP = "rctp"
)
const (
	CAND_HOST  = "host"
	CAND_SRFLX = "srflx"
	CAND_RELAY = "relay"
)

const (
	CAND_TCP_ACTIVE = "active"
)

const (
	MEDIA_UNKNOWN   = ""
	MEDIA_AUDIO     = "m=audio"
	MEDIA_VIDEO     = "m=video"
	MEDIA_DATA_CHAN = "m=application"
)

const (
	DTLS_SHA256 = "sha-256"
)
