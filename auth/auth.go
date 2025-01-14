// https://github.com/twitchdev/authentication-go-sample/blob/main/oauth-authorization-code/main.go

/*
Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License").
You may not use this file except in compliance with the License.
A copy of the License is located at
	http://aws.amazon.com/apache2.0/
or in the "license" file accompanying this file.
This file is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and limitations under the License.
*/

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"

	//"os"
	//"stream-chatbot/common"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

const (
	stateCallbackKey = "oauth-state-callback"
	oauthSessionName = "oauth-session"
	oauthTokenKey    = "oauth-token"
)

var (
	chatbotCreds map[string]string
	tokenChan    chan string
	clientID     string
	clientSecret string
	scopes       = []string{"chat:read", "chat:edit", "channel:manage:polls"}
	redirectURL  = "http://localhost:8080/redirect"
	oauth2Config *oauth2.Config
	// Generate a random cookieSecret on each run
	cookieSecret = []byte(generateRandomString(27))
	cookieStore  = sessions.NewCookieStore(cookieSecret)
)

func generateRandomString(length int) string {
	// Calculate the number of bytes needed for the desired string length
	bytes := make([]byte, (length + 1)) // +1 to handle odd lengths

	// Read random bytes from the crypto/rand source
	if _, err := rand.Read(bytes); err != nil {
		log.Println("Error generating the random string")
		return "ThisIsNotAGoodCookieSecret"
	}

	// Encode the bytes to a base64 string
	encoded := base64.StdEncoding.EncodeToString(bytes)

	return encoded[:length]
}

// HandleRoot is a Handler that shows a login button. In production, if the frontend is served / generated
// by Go, it should use html/template to prevent XSS attacks.
func HandleRoot(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html><body><a href="/login">Login using Twitch</a></body></html>`))

	return
}

// HandleLogin is a Handler that redirects the user to Twitch for login, and provides the 'state'
// parameter which protects against login CSRF.
func HandleLogin(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := cookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		err = nil
	}

	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return AnnotateError(err, "Couldn't generate a session!", http.StatusInternalServerError)
	}

	state := hex.EncodeToString(tokenBytes[:])

	session.AddFlash(state, stateCallbackKey)

	if err = session.Save(r, w); err != nil {
		return
	}

	http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)

	return
}

// HandleOauth2Callback is a Handler for oauth's 'redirect_uri' endpoint;
// it validates the state token and retrieves an OAuth token from the request parameters.
func HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := cookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		err = nil
	}

	// ensure we flush the csrf challenge even if the request is ultimately unsuccessful
	defer func() {
		if err := session.Save(r, w); err != nil {
			log.Printf("error saving session: %s", err)
		} else {
			log.Println("session saved via defer func")
		}
	}()

	switch stateChallenge, state := session.Flashes(stateCallbackKey), r.FormValue("state"); {
	case state == "", len(stateChallenge) < 1:
		err = errors.New("missing state challenge")
	case state != stateChallenge[0]:
		err = fmt.Errorf("invalid oauth state, expected '%s', got '%s'\n", state, stateChallenge[0])
		if len(stateChallenge) > 1 {
			log.Printf("multiple state challenges present: %v total\n", len(stateChallenge))
			/* for i, v := range stateChallenge {
				log.Printf("state challenge %d: %s\n", i, v)
			}
			log.Printf("state from request: %s\n", state) */
		}
	}

	if err != nil {
		return AnnotateError(
			err,
			"Couldn't verify your confirmation, please try again.",
			http.StatusBadRequest,
		)
	}

	token, err := oauth2Config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		return
	}

	// add the oauth token to session
	session.Values[oauthTokenKey] = token

	// Print the last 5 characters of the access token for debugging
	log.Printf("Access token from auth.go: %s\n", token.AccessToken[len(token.AccessToken)-5:])
	//fmt.Printf("Access token: %s\n", token.AccessToken)
	//fmt.Printf("Full token value: %v\n", token)

	go func() {
		tokenChan <- token.AccessToken
	}()
	log.Println("Sent token to channel")

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	log.Println("Redirected to /")

	return
}

// HumanReadableError represents error information
// that can be fed back to a human user.
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type HumanReadableError interface {
	HumanError() string
	HTTPCode() int
}

// HumanReadableWrapper implements HumanReadableError
type HumanReadableWrapper struct {
	ToHuman string
	Code    int
	error
}

func (h HumanReadableWrapper) HumanError() string { return h.ToHuman }
func (h HumanReadableWrapper) HTTPCode() int      { return h.Code }

// AnnotateError wraps an error with a message that is intended for a human end-user to read,
// plus an associated HTTP error code.
func AnnotateError(err error, annotation string, code int) error {
	if err == nil {
		return nil
	}
	return HumanReadableWrapper{ToHuman: annotation, error: err}
}

type Handler func(http.ResponseWriter, *http.Request) error

func TwitchAuth(TokenChan chan string, ChatbotCreds map[string]string) {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})

	chatbotCreds = ChatbotCreds
	clientID = chatbotCreds["ClientID"]
	clientSecret = chatbotCreds["ClientSecret"]

	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint:     twitch.Endpoint,
		RedirectURL:  redirectURL,
	}

	var middleware = func(h Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			// parse POST body, limit request size
			if err = r.ParseForm(); err != nil {
				return AnnotateError(err, "Something went wrong! Please try again.", http.StatusBadRequest)
			}

			return h(w, r)
		}
	}

	// errorHandling is a middleware that centralises error handling.
	// this prevents a lot of duplication and prevents issues where a missing
	// return causes an error to be printed, but functionality to otherwise continue
	// see https://blog.golang.org/error-handling-and-go
	var errorHandling = func(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := handler(w, r); err != nil {
				var errorString string = "Something went wrong! Please try again."
				var errorCode int = 500

				if v, ok := err.(HumanReadableError); ok {
					errorString, errorCode = v.HumanError(), v.HTTPCode()
				}

				log.Println(err)
				w.Write([]byte(errorString))
				w.WriteHeader(errorCode)
				return
			}
		})
	}

	var handleFunc = func(path string, handler Handler) {
		http.Handle(path, errorHandling(middleware(handler)))
	}

	log.Printf("Client ID: %s\n", clientID)
	tokenChan = TokenChan
	handleFunc("/", HandleRoot)
	handleFunc("/login", HandleLogin)
	handleFunc("/redirect", HandleOAuth2Callback)

	log.Println("Started running auth on http://localhost:8080/")
	fmt.Println("Open http://localhost:8080/ to authenticate with Twitch")
	log.Println(http.ListenAndServe(":8080", nil))
}
