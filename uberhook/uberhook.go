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

package uberhook

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/orijtech/authmid"
	"github.com/orijtech/uber/oauth2"
	"github.com/orijtech/uber/v1"
)

type Event struct {
	ID       string `json:"event_id"`
	TimeUnix int64  `json:"event_time"`
	Type     string `json:"event_type"`

	Meta *Meta `json:"meta"`

	URL string `json:"resource_href"`
}

type Status string

type Meta struct {
	UserID     string      `json:"user_id"`
	ResourceID string      `json:"resource_id"`
	Status     uber.Status `json:"status"`
}

type Webhook struct {
	sync.RWMutex
	oauthConfig *oauth2.OAuth2AppConfig
}

var _ authmid.Authenticator = (*Webhook)(nil)

var (
	errBlankClientSecret = errors.New("expecting a non-blank clientSecret")
)

func (v *Webhook) HeaderValues(hdr http.Header) ([]string, []string, error) {
	return nil, nil, nil
}

func (v *Webhook) LookupAPIKey(hdr http.Header) (string, error) {
	// Uber doesn't include the APIKey as part of the header signatures
	// at least as of `Sat 10 Jun 2017 01:20:15 MDT` so  send back a blank.
	// Reference: https://developer.uber.com/docs/riders/guides/webhooks
	// has no mention of client_id in there, only client_secret
	return "", nil
}

func (v *Webhook) LookupSecret(apiKey string) ([]byte, error) {
	var err error = errBlankClientSecret
	var secret []byte
	if v != nil {
		v.RLock()
		if v.oauthConfig != nil {
			secret = []byte(v.oauthConfig.ClientSecret)
			err = nil
		}
		v.RUnlock()
	}
	return secret, err
}

func (v *Webhook) Signature(hdr http.Header) (string, error) {
	return hdr.Get("X-Uber-Signature"), nil
}

func New() (*Webhook, error) {
	oauth2Config, err := oauth2.OAuth2ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return &Webhook{oauthConfig: oauth2Config}, nil
}

func (w *Webhook) Middleware(next http.Handler) http.Handler {
	return authmid.Middleware(w, next)
}

var _ authmid.ExcludeMethodAndPather = (*Webhook)(nil)

// Uber's webhook signature verification only consists of (clientSecret, webhookBody)
func (w *Webhook) ExcludeMethodAndPath() bool { return true }

var blankEvent Event
var (
	errBlankEvent = errors.New("expecting a non-blank event")
)

func FparseEvent(r io.Reader) (*Event, error) {
	blob, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ev := new(Event)
	if err := json.Unmarshal(blob, ev); err != nil {
		return nil, err
	}
	if *ev == blankEvent {
		return nil, errBlankEvent
	}
	return ev, nil
}
