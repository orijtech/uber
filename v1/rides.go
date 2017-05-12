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

	"github.com/orijtech/otils"
)

type RideRequest struct {
	FareID         string    `json:"fare_id,omitempty"`
	StartLatitude  float32   `json:"start_latitude,omitempty"`
	StartLongitude float32   `json:"start_longitude,omitempty"`
	StartPlace     PlaceName `json:"start_place_id,omitempty"`
	StartAddress   string    `json:"start_address,omitempty"`

	EndLatitude  float32   `json:"end_latitude,omitempty"`
	EndLongitude float32   `json:"end_longitude,omitempty"`
	EndPlace     PlaceName `json:"end_place_id,omitempty"`
	EndAddress   string    `json:"end_address,omitempty"`

	// Optional fields
	// PaymentMethodID if set will be the payment
	// that will be used for the user's trip otherwise,
	// the user's last payment will be used.
	PaymentMethodID string `json:"payment_method_id,omitempty"`

	// RideTypeID if set will be the product type used, otherwise
	// it will fallback to the cheapest ride option present.
	RideTypeID string `json:"product_id,omitempty"`

	// The Unique identifier of surge session for
	// a user. Required when returned from a 409
	// conflict response a previous POST attempt.
	SurgeConfirmationID string `json:"surge_confirmation_id"`
}

type RideResponse struct {
	RequestID  string  `json:"request_id"`
	ETAMinutes float64 `json:"eta"`

	SurgeMultiplier otils.NullableFloat64 `json:"surge_multiplier,omitempty"`
}

type RideConflict struct {
	SurgeConfirmationID string                `json:"surge_confirmation_id"`
	SurgeAcceptanceURL  string                `json:"href"`
	SurgeMultiplier     otils.NullableFloat64 `json:"multiplier"`

	// ExpiresAt is the Unix timestamp at which
	// the SurgeConfirmationID expires.
	ExpiresAt int64 `json:"expires_at"`
}

var (
	errEmptyFareID = errors.New("expecting a non-empty fareID")

	errInvalidRideRequest = errors.New("expecting (StartLatitude,EndLatitude: EndLatitude,EndLongitude) or (StartID, EndPlaceID) or (StartPlaceAddress, EndPlaceAddress)")
)

func (rr *RideRequest) Validate() error {
	if rr.FareID == "" {
		return errEmptyFareID
	}
	if rr.StartLatitude != 0.0 && rr.EndLatitude != 0.0 {
		return nil
	}
	if rr.EndAddress != "" && rr.StartAddress != "" {
		return nil
	}
	if rr.StartPlace != "" && rr.EndPlace != "" {
		return nil
	}

	return errInvalidRideRequest
}

var errUnimplemented = errors.New("unimplemented")

func (c *Client) RequestRide(rr *RideRequest) (*RideResponse, error) {
	if err := rr.Validate(); err != nil {
		return nil, err
	}

	blob, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/requests", baseURL)
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}

	// RequestRide is a delicate endpoint that has a diverse set
	// of errors and actionable ones for that matter.
	// For example:
	// * Surge conflict so the user has to accept
	//  the surge with a confirmation.
	slurp, _, err := c.doHTTPReq(req)
	if err != nil {
		return nil, err
	}
	rres := new(RideResponse)
	if err := json.Unmarshal(slurp, rres); err != nil {
		return nil, err
	}
	return rres, nil
}
