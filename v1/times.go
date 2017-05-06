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

type TimeEstimate struct {
	// Expected Time of Arrival for the product (in seconds).
	ETASeconds otils.NullableFloat64 `json:"estimate"`

	// Unique identifier representing a specific
	// product for a given longitude and latitude.
	// For example, uberX in San Francisco will have
	// a different ProductID than uberX in Los Angelese.
	ProductID string `json:"product_id"`

	// Display name of product.
	Name string `json:"display_name"`

	// Localized display name of product.
	LocalizedName string `json:"localized_display_name"`

	LimitPerPage int64 `json:"limit"`
}

var errNilTimeEstimateRequest = errors.New("expecting a non-nil timeEstimateRequest")

type TimeEstimatesPage struct {
	Estimates []*TimeEstimate `json:"times"`

	Count int64 `json:"count,omitempty"`

	Err        error
	PageNumber uint64
}

var timeExcludedValues = map[string]bool{
	"seat_count": true,
}

func (c *Client) EstimateTime(treq *EstimateRequest) (pagesChan chan *TimeEstimatesPage, cancelPaging func(), err error) {
	if treq == nil {
		return nil, nil, errNilTimeEstimateRequest
	}

	pager := new(Pager)
	if treq != nil {
		*pager = treq.Pager
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

	cancelChan, cancelFn := makeCancelParadigm()
	estimatesPageChan := make(chan *TimeEstimatesPage)
	go func() {
		defer close(estimatesPageChan)

		throttleDuration := 150 * time.Millisecond
		pageNumber := uint64(0)

		canPage := true
		for canPage {
			tp := new(TimeEstimatesPage)
			tp.PageNumber = pageNumber

			qv, err := otils.ToURLValues(treq)
			if err != nil {
				tp.Err = err
				estimatesPageChan <- tp
				return
			}

			fullURL := fmt.Sprintf("%s/estimates/time?%s", baseURL, qv.Encode())
			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				tp.Err = err
				estimatesPageChan <- tp
				return
			}

			slurp, _, err := c.doAuthAndHTTPReq(req)
			if err != nil {
				tp.Err = err
				estimatesPageChan <- tp
				return
			}

			if err := json.Unmarshal(slurp, tp); err != nil {
				tp.Err = err
				estimatesPageChan <- tp
				return
			}

			estimatesPageChan <- tp

			if tp.Count <= 0 {
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

	return estimatesPageChan, cancelFn, nil
}
