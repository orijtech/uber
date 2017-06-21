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
	"time"

	"github.com/orijtech/otils"
)

type EstimateRequest struct {
	StartLatitude  float64 `json:"start_latitude"`
	StartLongitude float64 `json:"start_longitude"`
	EndLongitude   float64 `json:"end_longitude"`
	EndLatitude    float64 `json:"end_latitude"`

	SeatCount int `json:"seat_count"`

	// ProductID is the UniqueID of the product
	// being requested. If unspecified, it will
	// default to the cheapest product for the
	// given location.
	ProductID string `json:"product_id"`

	StartPlace PlaceName `json:"start_place_id"`
	EndPlace   PlaceName `json:"end_place_id"`

	Pager
}

type PriceEstimate struct {
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

type PriceEstimatesPage struct {
	Estimates []*PriceEstimate `json:"prices"`

	Count int64 `json:"count,omitempty"`

	Err        error
	PageNumber uint64
}

func (c *Client) EstimatePrice(ereq *EstimateRequest) (pagesChan chan *PriceEstimatesPage, cancelPaging func(), err error) {
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

	cancelChan, cancelFn := makeCancelParadigm()
	estimatesPageChan := make(chan *PriceEstimatesPage)
	go func() {
		defer close(estimatesPageChan)

		throttleDuration := 150 * time.Millisecond
		pageNumber := uint64(0)

		canPage := true
		for canPage {
			ep := new(PriceEstimatesPage)
			ep.PageNumber = pageNumber

			qv, err := otils.ToURLValues(ereq)
			if err != nil {
				ep.Err = err
				estimatesPageChan <- ep
				return
			}

			fullURL := fmt.Sprintf("%s/estimates/price?%s", c.baseURL(), qv.Encode())
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

	return estimatesPageChan, cancelFn, nil
}

type FareEstimate struct {
	SurgeConfirmationURL string `json:"surge_confirmation_href,omitempty"`
	SurgeConfirmationID  string `json:"surge_confirmation_id"`

	// Breakdown provides details on how a fare came to be.
	Breakdown []*FareBreakdown `json:"fare_breakdown,omitempty"`

	SurgeMultiplier otils.NullableFloat64 `json:"surge_multiplier"`

	CurrencyCode  otils.NullableString `json:"currency_code"`
	DisplayAmount otils.NullableString `json:"display"`
}

type FareBreakdown struct {
	Low           otils.NullableFloat64 `json:"low_amount"`
	High          otils.NullableFloat64 `json:"high_amount"`
	DisplayAmount otils.NullableString  `json:"display_amount"`
	DisplayName   otils.NullableString  `json:"display_name"`
}

type Fare struct {
	Value         otils.NullableFloat64 `json:"value,omitempty"`
	ExpiresAt     int64                 `json:"expires_at,omitempty"`
	CurrencyCode  otils.NullableString  `json:"currency_code"`
	DisplayAmount otils.NullableString  `json:"display"`
	ID            otils.NullableString  `json:"fare_id"`
}

type UpfrontFare struct {
	Trip *Trip `json:"trip,omitempty"`
	Fare *Fare `json:"fare,omitempty"`

	// PickupEstimateMinutes is the estimated time of vehicle arrival
	// in minutes. It is unset if there are no cars available.
	PickupEstimateMinutes otils.NullableFloat64 `json:"pickup_estimate,omitempty"`

	Estimate *FareEstimate `json:"estimate,omitempty"`
}

func (upf *UpfrontFare) SurgeInEffect() bool {
	return upf != nil && upf.Estimate != nil && upf.Estimate.SurgeConfirmationURL != ""
}

func (upf *UpfrontFare) NoCarsAvailable() bool {
	return upf == nil || upf.PickupEstimateMinutes <= 0
}

var errInvalidSeatCount = errors.New("invalid seatcount, default and maximum value is 2")

func (esReq *EstimateRequest) validateForUpfrontFare() error {
	if esReq == nil {
		return errNilEstimateRequest
	}

	// The number of seats required for uberPool.
	// Default and maximum value is 2.
	if esReq.SeatCount < 0 || esReq.SeatCount > 2 {
		return errInvalidSeatCount
	}

	// UpfrontFares require:
	// * StartPlace or (StartLatitude, StartLongitude)
	// * EndPlace	or (EndLatitude,   EndLongitude)
	if esReq.StartPlace != "" && esReq.EndPlace != "" {
		return nil
	}

	// However, checks for unspecified zero floats require
	// special attention, so we'll let the JSON marshaling
	// serialize them.
	return nil
}

var errNilFare = errors.New("failed to unmarshal the response fare")

func (c *Client) UpfrontFare(esReq *EstimateRequest) (*UpfrontFare, error) {
	if err := esReq.validateForUpfrontFare(); err != nil {
		return nil, err
	}

	blob, err := json.Marshal(esReq)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/requests/estimate", c.baseURL())
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	slurp, _, err := c.doHTTPReq(req)
	if err != nil {
		return nil, err
	}

	upfrontFare := new(UpfrontFare)
	var blankUFare UpfrontFare
	if err := json.Unmarshal(slurp, upfrontFare); err != nil {
		return nil, err
	}
	if *upfrontFare == blankUFare {
		return nil, errNilFare
	}
	return upfrontFare, nil
}
