// Copyright 2017 orijtech. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type OAuth2AppConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

var (
	envOAuth2ClientIDKey     = "UBER_APP_OAUTH2_CLIENT_ID"
	envOAuth2ClientSecretKey = "UBER_APP_OAUTH2_CLIENT_SECRET"
)

func Transport(token *oauth2.Token) *oauth2.Transport {
	// Once we have the token we can now make the TokenSource
	ts := &tokenSourcer{token: token}
	return &oauth2.Transport{Source: ts}
}

func TransportWithBase(token *oauth2.Token, base http.RoundTripper) *oauth2.Transport {
	tr := Transport(token)
	tr.Base = base
	return tr
}

var (
	errNoOAuth2TokenDeserialized = errors.New("unable to deserialize an OAuth2.0 token")

	blankOAuth2Token = oauth2.Token{}
)

func TransportFromFile(path string) (*oauth2.Transport, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	token := new(oauth2.Token)
	if err := json.Unmarshal(blob, token); err != nil {
		return nil, err
	}
	if reflect.DeepEqual(blankOAuth2Token, *token) {
		return nil, errNoOAuth2TokenDeserialized
	}

	return Transport(token), nil
}

type tokenSourcer struct {
	sync.RWMutex
	token *oauth2.Token
}

var _ oauth2.TokenSource = (*tokenSourcer)(nil)

func (ts *tokenSourcer) Token() (*oauth2.Token, error) {
	ts.RLock()
	defer ts.RUnlock()

	return ts.token, nil
}

const (
	OAuth2AuthURL  = "https://login.uber.com/oauth/v2/authorize"
	OAuth2TokenURL = "https://login.uber.com/oauth/v2/token"
)

// OAuth2ConfigFromEnv retrieves your app's client id and client
// secret from your environment, with the purpose of later being
// able to perform application functions on behalf of users.
func OAuth2ConfigFromEnv() (*OAuth2AppConfig, error) {
	var errsList []string
	oauth2ClientID := strings.TrimSpace(os.Getenv(envOAuth2ClientIDKey))
	if oauth2ClientID == "" {
		errsList = append(errsList, fmt.Sprintf("%q was not set", envOAuth2ClientIDKey))
	}
	oauth2ClientSecret := strings.TrimSpace(os.Getenv(envOAuth2ClientSecretKey))
	if oauth2ClientSecret == "" {
		errsList = append(errsList, fmt.Sprintf("%q was not set", envOAuth2ClientSecretKey))
	}

	if len(errsList) > 0 {
		return nil, errors.New(strings.Join(errsList, "\n"))
	}

	config := &OAuth2AppConfig{
		ClientID:     oauth2ClientID,
		ClientSecret: oauth2ClientSecret,
	}

	return config, nil
}

const (
	// Access the user's basic profile information
	// on a user's Uber account including their
	// firstname, email address and profile picture.
	ScopeProfile = "profile"

	// Pull trip data including times, product
	// type andd city information of a user's
	// historical pickups and drop-offs.
	ScopeHistory = "history"

	// ScopeHistoryLite is the same as
	// ScopeHistory but without city information.
	ScopeHistoryLite = "history_lite"

	// Access to get and update your saved places.
	// This includes your home and work addresses if
	// you have saved them with Uber.
	ScopePlaces = "places"

	// Allows developers to provide a complete
	// Uber ride experience inside their app
	// using the widget. Enables users to access
	// trip information for rides requested through
	// the app and the current ride, available promos,
	// and payment methods (last two digits only)
	// using the widget. Uber's charges, terms
	// and policies will apply.
	ScopeRideWidgets = "ride_widgets"

	// ScopeRequest is a privileged scope that
	// allows your application to make requests
	// for Uber products on behalf of users.
	ScopeRequest = "request"

	// ScopeRequestReceipt is a privileged scope that
	// allows your application to get receipt details
	// for requests made by the application.
	// Restrictions: This scope is only granted to apps
	//  that request Uber rides directly and receipts
	//  as part of the trip lifecycle. We do not allow
	// apps to aggregate receipt information. The receipt
	// endpoint will only provide receipts for ride requests
	// origination from your application. It is not
	// currently possible to receive receipt
	// data for all trips, as of: (Fri 12 May 2017 18:18:42 MDT).
	ScopeRequestReceipt = "request_receipt"

	// ScopeAllTrips is a privileged scope that allows
	// access to trip details about all future Uber trips,
	// including pickup, destination and real-time
	// location for all of your future rides.
	ScopeAllTrips = "all_trips"
)

func AuthorizeByEnvApp(scopes ...string) (*oauth2.Token, error) {
	oconfig, err := OAuth2ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return Authorize(oconfig, scopes...)
}

func Authorize(oconfig *OAuth2AppConfig, scopes ...string) (*oauth2.Token, error) {
	config := &oauth2.Config{
		ClientID:     oconfig.ClientID,
		ClientSecret: oconfig.ClientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  OAuth2AuthURL,
			TokenURL: OAuth2TokenURL,
		},
	}

	state := fmt.Sprintf("%v%s", time.Now().Unix(), rand.Float32())
	urlToVisit := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Please visit this URL for the auth dialog: %v\n", urlToVisit)

	callbackURLChan := make(chan url.Values)
	go func() {
		http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			query := req.URL.Query()
			callbackURLChan <- query
			fmt.Fprintf(rw, "Received the token successfully. Please return to your terminal")
		})

		defer close(callbackURLChan)
		addr := ":8889"
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	urlValues := <-callbackURLChan
	gotState, wantState := urlValues.Get("state"), state
	if gotState != wantState {
		return nil, fmt.Errorf("states do not match: got: %q want: %q", gotState, wantState)
	}
	code := urlValues.Get("code")

	ctx := context.Background()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}
