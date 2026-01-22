package railzwayclient

import (
	"fmt"
)

type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("valora api error (%d): %s", e.Status, e.Message)
}


