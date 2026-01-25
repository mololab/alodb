package websocket

import "errors"

var (
	ErrQueryTimeout       = errors.New("query execution timed out")
	ErrClientDisconnected = errors.New("client disconnected")
)
