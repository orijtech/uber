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
	"strings"
)

type RideRequest struct {
	// FareID is the ID of the upfront fare. If FareID is blank
	// and you would like an inspection of current estimates,
	// set PromptOnFare to review the upfront fare.
	FareID string `json:"fare_id,omitempty"`

	// PromptOnFare is an optional callback function that is
	// used when FareID is blank. It is invoked to inspect and
	// accept the upfront fare estimate or any surges in effect.
	PromptOnFare func(*UpfrontFare) error `json:"-"`

	// StartPlace can be used in place of (StartLatitude, StartLongitude)
	StartPlace PlaceName `json:"start_place_id,omitempty"`

	// EndPlace can be used in place of (EndLatitude, EndLongitude)
	EndPlace PlaceName `json:"end_place_id,omitempty"`

	StartLatitude  float64 `json:"start_latitude,omitempty"`
	StartLongitude float64 `json:"start_longitude,omitempty"`
	EndLatitude    float64 `json:"end_latitude,omitempty"`
	EndLongitude   float64 `json:"end_longitude,omitempty"`

	// Optional fields
	// Product is the ID of the product being requested. If none is provided,
	// it will default to the cheapest product for the given location.
	ProductID string `json:"product_id,omitempty"`

	// SurgeConfirmationID is the unique identifier of the surge session for a user.
	// Required when returned from a 409 Conflict repsonse on a previous POST attempt.
	SurgeConfirmationID string `json:"surge_confirmation_id,omitempty"`

	// PaymentMethodID is the unique identifier of the payment method selected by a user.
	// If set, the trip will be requested using this payment method. If not set, the trip
	// will be requested using the user's last used payment method.
	PaymentMethodID string `json:"payment_method_id,omitempty"`

	// uberPOOL data
	// SeatCount is the number of seats required for uberPOOL.
	// The default and maximum value is 2.
	SeatCount int `json:"seat_count,omitempty"`

	// Uber for Business data
	// ExpenseCode is an alphanumeric identifier for expense reporting policies.
	// This value will appear in the trip receipt and any configured expense-reporting
	// integrations like:
	// * Uber For Business: https://www.uber.com/business
	// * Business Profiles: https://www.uber.com/business/profiles
	ExpenseCode string `json:"expense_code,omitempty"`

	// ExpenseMemo is a free text field to describe the purpose of the trip for
	// expense reporting. This value will appear in the trip receipt and any
	// configured expense-reporting integrations like:
	// * Uber For Business: https://www.uber.com/business
	// * Business Profiles: https://www.uber.com/business/profiles
	ExpenseMemo string `json:"expense_memo,omitempty"`
}

func (c *Client) preprocessBeforeValidate(rr *RideRequest) (*RideRequest, error) {
	if rr == nil || strings.TrimSpace(rr.FareID) != "" || rr.PromptOnFare == nil {
		return rr, nil
	}

	// Otherwise it is time to get the estimate of the fare
	upfrontFare, err := c.UpfrontFare(&EstimateRequest{
		StartLatitude:  rr.StartLatitude,
		StartLongitude: rr.StartLongitude,
		StartPlace:     rr.StartPlace,
		EndPlace:       rr.EndPlace,
		EndLatitude:    rr.EndLatitude,
		EndLongitude:   rr.EndLongitude,
		SeatCount:      rr.SeatCount,
	})

	if err != nil {
		return nil, err
	}

	modRreq := new(RideRequest)
	// Shallow copy of the original then modify the copy.
	*modRreq = *rr
	modRreq.FareID = string(upfrontFare.Fare.ID)
	if modRreq.ProductID == "" {
		modRreq.ProductID = upfrontFare.Trip.ProductID
	}

	// Otherwise prompt for acceptance
	if err := rr.PromptOnFare(upfrontFare); err != nil {
		return nil, err
	}

	return modRreq, nil
}

func (rr *RideRequest) Validate() error {
	if rr == nil || strings.TrimSpace(rr.FareID) == "" {
		return ErrInvalidFareID
	}

	// Either:
	// 1. Start:
	//    * StartPlace
	//    * (StartLatitude, StartLongitude)
	// 2. End:
	//    * EndPlace
	//    * (EndLatitude, EndLongitude)
	if blankPlaceOrCoords(rr.StartPlace, rr.StartLatitude, rr.StartLongitude) {
		return ErrInvalidStartPlaceOrCoords
	}

	if blankPlaceOrCoords(rr.EndPlace, rr.EndLatitude, rr.EndLongitude) {
		return ErrInvalidEndPlaceOrCoords
	}

	return nil
}

func (c *Client) RequestRide(rreq *RideRequest) (*Ride, error) {
	rr, err := c.preprocessBeforeValidate(rreq)
	if err != nil {
		return nil, err
	}

	if err := rr.Validate(); err != nil {
		return nil, err
	}

	blob, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}
	fullURL := fmt.Sprintf("%s/requests", c.baseURL())
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	blob, _, err = c.doAuthAndHTTPReq(req)
	if err != nil {
		return nil, err
	}
	ride := new(Ride)
	if err := json.Unmarshal(blob, ride); err != nil {
		return nil, err
	}
	return ride, nil
}

var (
	ErrInvalidStartPlaceOrCoords = errors.New("invalid startPlace or (startLat, startLon)")
	ErrInvalidEndPlaceOrCoords   = errors.New("invalid endPlace or (endLat, endLon)")
)

func blankPlaceOrCoords(place PlaceName, lat, lon float64) bool {
	if strings.TrimSpace(string(place)) != "" {
		switch place {
		case PlaceHome, PlaceWork:
			return false
		default:
			return true
		}
	}

	// Otherwise now check out the coordinates
	// Coordinates can be any value
	return false
}

type Ride struct {
	RequestID string `json:"request_id"`
	ProductID string `json:"product_id"`

	// Status indicates the state of the ride request.
	Status Status `json:"status"`

	Vehicle  *Vehicle  `json:"vehicle,omitempty"`
	Driver   *Driver   `json:"driver"`
	Location *Location `json:"location"`

	// ETAMinutes is the expected time of arrival in minutes.
	ETAMinutes int `json:"eta"`

	// The surge pricing multiplier used to calculate the increased price of a request.
	// A surge multiplier of 1.0 means surge pricing is not in effect.
	SurgeMultiplier float32 `json:"surge_multiplier"`
}

func (r *Ride) SurgeInEffect() bool {
	return r != nil && r.SurgeMultiplier == 1.0
}

type Vehicle struct {
	Model string `json:"model"`
	Make  string `json:"make"`

	LicensePlate string `json:"license_plate"`
	PictureURL   string `json:"picture_url"`
}

type Driver struct {
	PhoneNumber string `json:"phone_number"`
	SMSNumber   string `json:"sms_number"`

	PictureURL string `json:"picture_url"`
	Name       string `json:"name"`
	Rating     int    `json:"rating"`
}

type State string

type Location struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`

	// Bearing is the current bearing of the vehicle in degrees (0-359).
	Bearing int `json:"bearing,omitempty"`

	PrimaryAddress   string `json:"address,omitempty"`
	SecondaryAddress string `json:"address_2,omitempty"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	PostalCode       string `json:"postal_code,omitempty"`
	Country          string `json:"country,omitempty"`
}
