package callable

import (
	"encoding/json"
	"fmt"
	"net/url"
)

//Handler -
type Handler func(ctx Context) (interface{}, error)

type contextKey string

//Context -
type Context struct {
	Auth       Auth
	Data       interface{}
	InstanceID string
	URL        url.URL
}

func (it Context) String() string {
	return fmt.Sprintf("Context { Auth: %v, Data: %v, InstanceID: %v, URL: %v }",
		it.Auth, it.Data, it.InstanceID, it.URL)
}

//GetData - returns the data
func (it Context) GetData(data interface{}) error {
	if it.Data == nil {
		return fmt.Errorf("GetData: Data is nil")
	} else if jsonData, err := json.Marshal(it.Data); err != nil {
		return fmt.Errorf("GetData: %v", err)
	} else if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("GetData: %v", err)
	}
	return nil
}

//Auth -
type Auth struct {
	UID string
}

func (it Auth) String() string {
	return fmt.Sprintf("Auth { UID: %v }", it.UID)
}

type httpsError struct {
	Code          string         `json:"code"`
	Details       string         `json:"details"`
	HTTPErrorCode httpsErrorCode `json:"httpErrorCode"`
}

type httpsErrorCode struct {
	CanonicalName string `json:"canonicalName"`
	Status        int    `json:"status"`
}

type httpsCallableResult struct {
	Error  httpsError  `json:"error"`
	Result interface{} `json:"result"`
}
