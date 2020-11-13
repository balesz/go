package callable

//Handler -
type Handler func(ctx Context) (interface{}, error)

type contextKey string

//Context -
type Context struct {
	Auth       Auth
	Data       interface{}
	InstanceID string
}

//Auth -
type Auth struct {
	UID string
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
