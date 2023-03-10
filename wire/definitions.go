package wire

import "time"

// algorithm in routing
const (
	AlgorithmHashSlots = "hashslots"
)

// Command defined data type betweem client and server
const (
	// login
	CommandLoginSignIn  = "login.signin"
	CommandLoginSignOut = "login.signout"

	// chat
	CommandChatUserTalk  = "chat.user.talk"
	CommandChatGroupTalk = "chat.group.talk"
	CommandChatTalkAck   = "chat.talk.ack"

	// offline
	CommandOfflineIndex   = "chat.offline.index"
	CommandOfflineContent = "chat.offline.content"

	// group
	CommandGroupCreate  = "chat.group.create"
	CommandGroupJoin    = "chat.group.join"
	CommandGroupQuit    = "chat.group.quit"
	CommandGroupMembers = "chat.group.members"
	CommandGroupDetail  = "chat.group.detail"
)

// Meta key of a packet
const (
	// ServiceName of the gateway the message will sent to
	MetaDestServer = "dest.server"
	// Channels the message will sent to
	MetaDestChannels = "dest.channels"
)

// Protocol
type Protocol string

const (
	ProtocolTCP       Protocol = "tcp"
	ProtocolWebsocket Protocol = "websocket"
)

// ServiceName
const (
	SNWGateway = "wgateway"
	SNTGateway = "tgateway"
	SNLogin    = "login"
	SNChat     = "chat"
	SNService  = "royal" // rpc service
)

type ServiceID string

type SessionID string

type Magic [4]byte

var (
	MagicLogicPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
	MagicBasicPkt = Magic{0xc3, 0x15, 0xa7, 0x65}
)

const (
	OfflineReadIndexExpiresIn = time.Hour * 24 * 30
	OfflineMessageExpiresIn   = 15
	OfflineSyncIndexCount     = 2000
	MessageMaxCountPerPage    = 200
)

const (
	MessageTypeText  = 1
	MessageTypeImage = 2
	MessageTypeVoice = 3
	MessageTypeVideo = 4
)
