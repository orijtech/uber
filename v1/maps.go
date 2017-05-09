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
	"errors"
	"fmt"
	"net/http"

	"github.com/skratchdot/open-golang/open"
)

type Map struct {
	RequestID string `json:"request_id"`

	URL string `json:"href"`
}

var (
	errEmptyTripID = errors.New("expecting a non-empty tripID")
	errNoSuchMap   = errors.New("no such map")
)

func (c *Client) RequestMap(tripID string) (*Map, error) {
	if tripID == "" {
		return nil, errEmptyTripID
	}

	fullURL := fmt.Sprintf("%s/requests/%s/map", baseURL, tripID)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	slurp, _, err := c.doAuthAndHTTPReq(req)
	if err != nil {
		return nil, err
	}

	uinfo := new(Map)
	blankMap := *uinfo
	if err := json.Unmarshal(slurp, uinfo); err != nil {
		return nil, err
	}
	if blankMap == *uinfo {
		return nil, errNoSuchMap
	}
	return uinfo, nil
}

// OpenMapForTrip is a convenience method that opens the map
// for a trip or returns an error if it encounters an error.
func (c *Client) OpenMapForTrip(tripID string) error {
	uinfo, err := c.RequestMap(tripID)
	if err != nil {
		return err
	}
	return open.Start(uinfo.URL)
}
