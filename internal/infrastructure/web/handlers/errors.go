package handlers

type APIKeyError struct {
	Message string
}

func (e *APIKeyError) Error() string {
	return e.Message
}
