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
	"time"

	"github.com/orijtech/otils"
)

type EstimateRequest struct {
	StartLatitude  float64 `json:"start_latitude"`
	StartLongitude float64 `json:"start_longitude"`
	EndLongitude   float64 `json:"end_longitude"`
	EndLatitude    float64 `json:"end_latitude"`
	SeatCount      int     `json:"seat_count"`

	Pager
}

type Estimate struct {
	// ISO 4217 currency code.
	CurrencyCode otils.NullableString `json:"currency_code"`

	// Formatted string of estimate in local currency of the
	// start location. Estimate could be a range, a single
	// number(flat rate) or "Metered" for TAXI.
	Estimate otils.NullableString `json:"estimate"`

	// Expected activity duration in seconds.
	DurationSeconds otils.NullableFloat64 `json:"duration"`

	// Minimum price for product.
	MinimumPrice otils.NullableFloat64 `json:"minimum"`

	// Lower bound of the estimated price.
	LowEstimate otils.NullableFloat64 `json:"low_estimate"`

	// Upper bound of the estimated price.
	HighEstimate otils.NullableFloat64 `json:"high_estimate"`

	// Unique identifier representing a specific
	// product for a given longitude and latitude.
	// For example, uberX in San Francisco will have
	// a different ProductID than uberX in Los Angelese.
	ProductID string `json:"product_id"`

	// Display name of product.
	Name string `json:"display_name"`

	// Localized display name of product.
	LocalizedName string `json:"localized_display_name"`

	// Expected surge multiplier. Surge is active if
	// SurgeMultiplier is greater than 1. Price estimate
	// already factors in the surge multiplier.
	SurgeMultiplier otils.NullableFloat64 `json:"surge_multiplier"`

	LimitPerPage int64 `json:"limit"`
}

var errNilEstimateRequest = errors.New("expecting a non-nil estimateRequest")

type EstimatesPage struct {
	Estimates []*Estimate `json:"prices"`

	Count int64 `json:"count,omitempty"`

	Err        error
	PageNumber uint64
}

func (c *Client) EstimatePrice(ereq *EstimateRequest) (chan *EstimatesPage, chan bool, error) {
	if ereq == nil {
		return nil, nil, errNilEstimateRequest
	}

	pager := new(Pager)
	if ereq != nil {
		*pager = ereq.Pager
	}

	// Adjust the paging parameters since they'll be heavily used
	pager.adjustPageParams()

	requestedMaxPage := pager.MaxPages
	pageNumberExceeds := func(pageNumber uint64) bool {
		if requestedMaxPage <= 0 {
			// No page limit requested at all here
			return false
		}

		return pageNumber >= uint64(requestedMaxPage)
	}

	cancelChan := make(chan bool, 1)
	estimatesPageChan := make(chan *EstimatesPage)
	go func() {
		defer close(estimatesPageChan)

		throttleDuration := 150 * time.Millisecond
		pageNumber := uint64(0)

		canPage := true
		for canPage {
			ep := new(EstimatesPage)
			ep.PageNumber = pageNumber

			qv, err := otils.ToURLValues(ereq)
			if err != nil {
				ep.Err = err
				estimatesPageChan <- ep
				return
			}

			fullURL := fmt.Sprintf("%s/estimates/price?%s", baseURL, qv.Encode())
			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				ep.Err = err
				estimatesPageChan <- ep
				return
			}

			slurp, _, err := c.doAuthAndHTTPReq(req)
			if err != nil {
				ep.Err = err
				estimatesPageChan <- ep
				return
			}

			if err := json.Unmarshal(slurp, ep); err != nil {
				ep.Err = err
				estimatesPageChan <- ep
				return
			}

			estimatesPageChan <- ep

			if ep.Count <= 0 {
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

			pager.StartOffset += pager.LimitPerPage
		}
	}()

	return estimatesPageChan, cancelChan, nil
}
