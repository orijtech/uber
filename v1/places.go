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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type PlaceName string

const (
	PlaceHome PlaceName = "home"
	PlaceWork PlaceName = "work"
)

func (c *Client) Place(placeName PlaceName) (*Place, error) {
	fullURL := fmt.Sprintf("%s/places/%s", c.baseURL(), placeName)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	return c.doPlaceReq(req)
}

func (c *Client) doPlaceReq(req *http.Request) (*Place, error) {
	slurp, _, err := c.doReq(req)
	if err != nil {
		return nil, err
	}

	place := new(Place)
	if err := json.Unmarshal(slurp, place); err != nil {
		return nil, err
	}
	return place, nil
}

type PlaceParams struct {
	Place   PlaceName `json:"place"`
	Address string    `json:"address"`
}

var (
	errEmptyAddress     = errors.New("expecting a non-empty address")
	errInvalidPlaceName = fmt.Errorf("invalid placeName; can only be either %q or %q", PlaceHome, PlaceWork)
)

func (pp *PlaceParams) Validate() error {
	if pp == nil || pp.Address == "" {
		return errEmptyAddress
	}

	switch pp.Place {
	case PlaceHome, PlaceWork:
		return nil
	default:
		return errInvalidPlaceName
	}
}

// UpdatePlace udpates your place's address.
func (c *Client) UpdatePlace(pp *PlaceParams) (*Place, error) {
	if err := pp.Validate(); err != nil {
		return nil, err
	}

	blob, err := json.Marshal(&Place{Address: pp.Address})
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/places/%s", c.baseURL(), pp.Place)
	req, err := http.NewRequest("PUT", fullURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doPlaceReq(req)
}
