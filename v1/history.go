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
	"time"

	"github.com/orijtech/otils"
)

type Status string

type Trip struct {
	// Status of the activity. As per API v1.2,
	// it only return "completed" for now.
	Status Status `json:"status,omitempty"`

	// Length of activity in miles.
	DistanceMiles float64 `json:"distance,omitempty"`

	// UnixTimestamp of activity start time.
	StartTimeUnix int64 `json:"start_time,omitempty"`

	// UnixTimestamp of activity end time.
	EndTimeUnix int64 `json:"end_time,omitempty"`

	// The city in which this trip was initiated.
	StartCity *Place `json:"start_city,omitempty"`

	ProductID string `json:"product_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

type Place struct {
	// The latitude of the approximate city center.
	Latitude float64 `json:"latitude,omitempty"`

	Name string `json:"display_name,omitempty"`

	// The longitude of the approximate city center.
	Longitude float64 `json:"longitude,omitempty"`

	Address string `json:"address,omitempty"`
}

type TripThread struct {
	Trips  []*Trip `json:"history"`
	Count  int64   `json:"count"`
	Limit  int64   `json:"limit"`
	Offset int64   `json:"offset"`
}

type Pager struct {
	ThrottleDuration time.Duration `json:"-"`
	LimitPerPage     int64         `json:"limit"`
	MaxPages         int64         `json:"-"`
	StartOffset      int64         `json:"offset"`
}

type TripThreadPage struct {
	TripThread
	Err        error
	PageNumber uint64
}

const (
	DefaultLimitPerPage = int64(50)
	DefaultStartOffset  = int64(0)
)

func (treq *Pager) adjustPageParams() {
	if treq.LimitPerPage <= 0 {
		treq.LimitPerPage = DefaultLimitPerPage
	}
	if treq.StartOffset <= 0 {
		treq.StartOffset = DefaultStartOffset
	}
}

func (c *Client) ListAllMyHistory() (thChan chan *TripThreadPage, cancelFn func(), err error) {
	return c.ListHistory(nil)
}

func (c *Client) ListHistory(threq *Pager) (thChan chan *TripThreadPage, cancelFn func(), err error) {
	treq := new(Pager)
	if threq != nil {
		*treq = *threq
	}

	// Adjust the paging parameters since they'll be heavily used
	treq.adjustPageParams()

	requestedMaxPage := treq.MaxPages
	pageNumberExceeds := func(pageNumber uint64) bool {
		if requestedMaxPage <= 0 {
			// No page limit requested at all here
			return false
		}

		return pageNumber >= uint64(requestedMaxPage)
	}

	cancelChan, cancelFn := makeCancelParadigm()

	historyChan := make(chan *TripThreadPage)
	go func() {
		defer close(historyChan)

		throttleDuration := 150 * time.Millisecond
		pageNumber := uint64(0)

		canPage := true
		for canPage {
			ttp := new(TripThreadPage)
			ttp.PageNumber = pageNumber

			qv, err := otils.ToURLValues(treq)
			if err != nil {
				ttp.Err = err
				historyChan <- ttp
				return
			}

			fullURL := fmt.Sprintf("%s/history?%s", baseURL, qv.Encode())
			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				ttp.Err = err
				historyChan <- ttp
				return
			}

			slurp, _, err := c.doAuthAndHTTPReq(req)
			if err != nil {
				ttp.Err = err
				historyChan <- ttp
				return
			}

			if err := json.Unmarshal(slurp, ttp); err != nil {
				ttp.Err = err
				historyChan <- ttp
				return
			}

			historyChan <- ttp

			if ttp.Count <= 0 {
				// No more items to page
				return
			}

			select {
			case <-cancelChan:
				// The user has canceled the paging
				canPage = false
				return

			case <-time.After(throttleDuration):
				// Do nothing here, the throttle time expired.
			}

			// Now it is time to adjust the next offset accordingly with the remaining count

			// Increment the page number as well
			pageNumber += 1
			if pageNumberExceeds(pageNumber) {
				return
			}

			treq.StartOffset += treq.LimitPerPage
		}
	}()

	return historyChan, cancelFn, nil
}
