package msg

type MqttMsg struct {
	Topic string
	Msg string
}

type UserEvent int 
const (
	Reset 	= iota
    Login
    Logout
    Create
    Close
    ChooseNGHero
    Invite
    Join
    StartQueue
    CancelQueue
    UpdateGame
    PreStart
    PreStartGet
    Leave
    StartGame
    GameOver
    GameInfo
    GameClose
    Status
    Reconnect
    MainServerDead
)

type ServerMsg struct {
	Event UserEvent
	Id string
	Msg string
}

type SqlEvent int 
const (
	SqlLogin 	= iota
	UpdateScore
	UpdateGameInfo
)

type SqlMsg struct {
	Event SqlEvent
	Msg string
}

type QueueEvent int 
const (
	UpdateRoom 	= iota
	RemoveRoom
)

type QueueMsg struct {
	Event QueueEvent
	Msg string
}