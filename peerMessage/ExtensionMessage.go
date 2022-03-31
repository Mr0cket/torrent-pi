package message

type ExtendedMsgID uint8

const (
	ExtHandshake ExtendedMsgID = 0
)

type ExtendedMessage struct {
	Message
}

func New() {
	//1. uint32 length of msg

	// 2. uint8 bittorrent messageId
	// MsgExtended

}

// The metadata extension uses the extension protocol (specified in BEP 0010 ) to advertize its existence.
// It adds the "ut_metadata" entry to the "m" dictionary in the extension header hand-shake message.
// This identifies the message code used for this message.
// It also adds "metadata_size" to the handshake message (not the "m" dictionary)
// specifying an integer value of the number of bytes of the metadata.
