package callable

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/balesz/go/firebase"
	"github.com/balesz/go/firebase/functions/logging"
)

//ContextKey -
const ContextKey = contextKey("context")

//Initializer -
func Initializer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var auth Auth
		var data interface{}
		var httpsError httpsError

		var err error
		if data, err = validateRequest(r); err != nil {
			httpsError = newError("invalid-argument", "Bad Request", http.StatusBadRequest, err)
		} else if auth, err = authenticate(r); err != nil {
			httpsError = newError("unauthenticated", "Unauthenticated", http.StatusUnauthorized, err)
		}

		if httpsError.Code != "" {
			if result, err := json.Marshal(httpsCallableResult{Error: httpsError}); err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			} else {
				w.Header().Add("Content-Type", "application/json")
				w.Write(result)
				w.WriteHeader(200)
				return
			}
		}

		instanceID := r.Header.Get("Firebase-Instance-ID-Token")

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ContextKey, Context{
			Auth: auth, Data: data, InstanceID: instanceID,
		})))
	})
}

//NewHandler -
func NewHandler(handler Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(ContextKey)
		if result, err := handler(ctx.(Context)); err != nil {
			logging.Error(fmt.Errorf("Error: %v", err))
			httpsError := newError("internal", "INTERNAL", http.StatusInternalServerError, err)
			if result, err := json.Marshal(httpsCallableResult{Error: httpsError}); err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			} else {
				w.Header().Add("Content-Type", "application/json")
				w.Write(result)
				w.WriteHeader(200)
				return
			}
		} else if result != nil {
			logging.Info(fmt.Sprintf("Result: %v", result))
			if result, err := json.Marshal(httpsCallableResult{Result: result}); err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			} else {
				w.Header().Add("Content-Type", "application/json")
				w.Write(result)
				w.WriteHeader(200)
				return
			}
		} else {
			http.Error(w, "Unknown Error", http.StatusInternalServerError)
		}
	})
}

func validateRequest(r *http.Request) (data interface{}, err error) {
	if r.Method != "POST" {
		err = fmt.Errorf("Request has invalid method (%v)", r.Method)
		return
	}

	if val := r.Header.Get("Content-Type"); !strings.HasPrefix(strings.ToLower(val), "application/json") {
		err = fmt.Errorf("Request has incorrect Content-Type (%v)", val)
		return
	}

	var res struct {
		Data interface{} `json:"data"`
	}

	payload, err := ioutil.ReadAll(r.Body)

	if err != nil {
		err = fmt.Errorf("Request is missing body (%v)", err)
		return
	} else if err = json.Unmarshal(payload, &res); err != nil {
		err = fmt.Errorf("JSON unmarshal error (%v)", err)
		return
	} else if res.Data == nil {
		err = fmt.Errorf("Request body is missing data (%v)", string(payload))
		return
	}

	data = res.Data
	return
}

func authenticate(r *http.Request) (auth Auth, err error) {
	rxToken := regexp.MustCompile("^Bearer (.*)$")

	var idToken string
	if authorization := r.Header.Get("Authorization"); authorization == "" {
		err = fmt.Errorf("Authorization header not exists")
		return
	} else if !rxToken.MatchString(authorization) {
		err = fmt.Errorf("Authorization header is invalid")
		return
	} else {
		idToken = strings.TrimSpace(rxToken.FindStringSubmatch(authorization)[1])
	}

	token, err := firebase.Auth.VerifyIDToken(context.Background(), idToken)
	//userID, err := verifyIDToken(idToken)
	if err != nil {
		return
	}

	auth = Auth{UID: token.UID}
	return
}

func verifyIDToken(token string) (string, error) {
	type Result struct {
		Expiration int64  `json:"exp"`
		Project    string `json:"aud"`
		UserID     string `json:"user_id"`
	}
	var result Result
	parts := strings.Split(token, ".")
	if jsonStr, err := base64.StdEncoding.DecodeString(parts[1]); err != nil {
		return "", err
	} else if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "", err
	} else if result.Project == "" {
		return "", fmt.Errorf("The payload of the token is invalid")
	} else if result.UserID == "" {
		return "", fmt.Errorf("The payload of the token is invalid")
	} else if exp := time.Unix(result.Expiration, 0); time.Now().After(exp) {
		return "", fmt.Errorf("The token is expired")
	}
	return result.UserID, nil
}

func newError(code string, name string, status int, err error) httpsError {
	return httpsError{
		Code:          code,
		Details:       err.Error(),
		HTTPErrorCode: httpsErrorCode{CanonicalName: name, Status: status},
	}
}
