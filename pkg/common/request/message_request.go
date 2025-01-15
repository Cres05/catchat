package request

type MessageRequest struct {
	MessageType int32  `json:"messageType"`
	Account     string `json:"user account"`
	ToAccount   string `json:"to user account"`
}
