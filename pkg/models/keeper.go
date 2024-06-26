package models

type MessageType string
type ItemType string

func (t ItemType) String() string {
	return string(t)
}

func (t MessageType) String() string {
	return string(t)
}

const (
	Update   MessageType = "update"
	New      MessageType = "new"
	Snapshot MessageType = "snapshot"
	Error    MessageType = "error"
)

const (
	CredItem ItemType = "cred"
	TextItem ItemType = "text"
	BinItem  ItemType = "bin"
	CardItem ItemType = "card"
)

type Credentials struct {
	Type     ItemType `json:"type"` //cred
	Tag      string   `json:"tag"`
	Login    string   `json:"login"`
	Password string   `json:"password"`
	Comment  string   `json:"comment"`
	Created  int64    `json:"created"`
}

type Text struct {
	Type    ItemType `json:"type"` //text
	Tag     string   `json:"tag"`
	Key     string   `json:"key"`
	Value   string   `json:"value"`
	Comment string   `json:"comment"`
	Created int64    `json:"created"`
}

type Binary struct {
	Type    ItemType `json:"type"` //bin
	Tag     string   `json:"tag"`
	Key     string   `json:"key"`
	Value   []byte   `json:"value"`
	Comment string   `json:"comment"`
	Created int64    `json:"created"`
}

type Card struct {
	Type    ItemType `json:"type"` //card
	Tag     string   `json:"tag"`
	Number  string   `json:"number"`
	Exp     string   `json:"exp"`
	CVV     int32    `json:"cvv"`
	Comment string   `json:"comment"`
	Created int64    `json:"created"`
}

// message from server - snapshot, update, error
// message from client - new
type Message struct {
	Token string      `json:"token"`
	Type  MessageType `json:"type"`
	Value []byte      `json:"value"`
}
