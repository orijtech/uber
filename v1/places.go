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
	"net/http"
)

type Address string

const (
	AddressHome Address = "home"
	AddressWork Address = "work"
)

func (c *Client) Place(address Address) (*Place, error) {
	fullURL := fmt.Sprintf("%s/places/%s", baseURL, address)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	slurp, _, err := c.doAuthAndHTTPReq(req)
	if err != nil {
		return nil, err
	}

	place := new(Place)
	if err := json.Unmarshal(slurp, place); err != nil {
		return nil, err
	}
	return place, nil
}
