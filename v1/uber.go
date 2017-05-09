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

package uber

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/orijtech/otils"
)

const baseURL = "https://api.uber.com/v1.2"

type Client struct {
	sync.RWMutex

	rt    http.RoundTripper
	token string
}

const envUberTokenKey = "UBER_TOKEN_KEY"

var errUnsetTokenEnvKey = fmt.Errorf("could not find %q in your environment", envUberTokenKey)

func NewClient(tokens ...string) (*Client, error) {
	if token := otils.FirstNonEmptyString(tokens...); token != "" {
		return &Client{token: token}, nil
	}

	// Otherwise fallback to retrieving it from the environment
	return NewClientFromEnv()
}

func NewClientFromEnv() (*Client, error) {
	retrToken := strings.TrimSpace(os.Getenv(envUberTokenKey))
	if retrToken == "" {
		return nil, errUnsetTokenEnvKey
	}

	return &Client{token: retrToken}, nil

}

func (c *Client) SetHTTPRoundTripper(rt http.RoundTripper) {
	c.Lock()
	c.rt = rt
	c.Unlock()
}

func (c *Client) SetBearerToken(token string) {
	c.Lock()
	defer c.Unlock()

	c.token = token
}

func (c *Client) httpClient() *http.Client {
	c.RLock()
	rt := c.rt
	c.RUnlock()

	if rt == nil {
		rt = http.DefaultTransport
	}

	return &http.Client{Transport: rt}
}

func (c *Client) bearerToken() string {
	c.RLock()
	defer c.RUnlock()

	return fmt.Sprintf("Bearer %s", c.token)
}

func (c *Client) tokenToken() string {
	c.RLock()
	defer c.RUnlock()

	return fmt.Sprintf("Token %s", c.token)
}

func (c *Client) doAuthAndHTTPReq(req *http.Request) ([]byte, http.Header, error) {
	req.Header.Set("Authorization", c.bearerToken())
	res, err := c.httpClient().Do(req)
	if err != nil {
		return nil, nil, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	if !otils.StatusOK(res.StatusCode) {
		errMsg := res.Status
		var err error
		if res.Body != nil {
			slurp, _ := ioutil.ReadAll(res.Body)
			if len(slurp) > 0 {
				ue := new(Error)
				if jerr := json.Unmarshal(slurp, ue); jerr == nil {
					err = ue
				} else {
					errMsg = string(slurp)
				}
			}
		}
		if err == nil {
			err = otils.MakeCodedError(errMsg, res.StatusCode)
		}
		return nil, res.Header, err
	}

	blob, err := ioutil.ReadAll(res.Body)
	return blob, res.Header, err
}

func makeCancelParadigm() (<-chan bool, func()) {
	var cancelOnce sync.Once
	cancelChan := make(chan bool, 1)
	cancelFn := func() {
		cancelOnce.Do(func() {
			close(cancelChan)
		})
	}

	return cancelChan, cancelFn
}

type Error struct {
	Meta   interface{}         `json:"meta"`
	Errors []*statusCodedError `json:"errors"`

	memoized string
}

func (ue *Error) Error() string {
	if ue == nil {
		return ""
	}
	if ue.memoized != "" {
		return ue.memoized
	}

	// Otherwise create it
	var errsList []string
	for _, sce := range ue.Errors {
		errsList = append(errsList, sce.Error())
	}
	ue.memoized = strings.Join(errsList, "\n")
	return ue.memoized
}

var _ error = (*Error)(nil)
var _ error = (*statusCodedError)(nil)

type statusCodedError struct {
	// The json tags are intentionally reversed
	// because an uber status coded error looks
	// like this:
	// {
	//    "status":404,
	//    "code":"unknown_place_id",
	//    "title":"Could not resolve the given place_id."
	// }
	// of which the above definitions seem reversed compared to
	// Go's net/http Request where Status is a message and StatusCode is an int.
	Code    int    `json:"status"`
	Message string `json:"code"`
	Title   string `json:"title"`

	memoizedErr string
}

func (sce *statusCodedError) Error() string {
	if sce == nil {
		return ""
	}
	if sce.memoizedErr == "" {
		blob, _ := json.Marshal(sce)
		sce.memoizedErr = string(blob)
	}
	return sce.memoizedErr
}
