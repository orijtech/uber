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

type DeliveryRequest struct {
	// The ID of the quoted price of the
	// delivery. This field is optional.
	// If missing, the fee for the delivery will
	// be determined at the time of the request.
	QuoteID string `json:"quote_id,omitempty"`

	// The merchant supplied order reference identifier.
	// This field is optional and it is limited to 256 characters.
	OrderReferenceID string `json:"order_reference_id,omitempty"`

	// The items being delivered.
	Items []*Item `json:"items"`

	// The details of the delivery pickup.
	Pickup  *Endpoint `json:"pickup"`
	Dropoff *Endpoint `json:"dropoff"`
}

type Item struct {
	Title    string `json:"title"`
	Fragile  bool   `json:"is_fragile,omitempty"`
	Quantity int    `json:"quantity"`

	WidthInches  float32 `json:"width,omitempty"`
	HeightInches float32 `json:"height,omitempty"`
	LengthInches float32 `json:"length,omitempty"`

	CurrencyCode CurrencyCode `json:"currency_code,omitempty"`
}

type Endpoint struct {
	Location *Location `json:"location"`
	Contact  *Contact  `json:"contact,omitempty"`

	// Special instructions for the endpoint. This field
	// is optional and it is limited to 256 characters.
	SpecialInstructions otils.NullableString `json:"special_instructions,omitempty"`

	SignatureRequired bool `json:"signature_required"`

	// Indicates if the delivery includes alcohol. This
	// feature is only available to whitelisted businesses.
	IncludesAlcohol bool `json:"includes_alcohol"`

	ETAMinutes int `json:"eta"`
}

type Contact struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       *Phone `json:"phone,omitempty"`

	// SendEmailNotifications if set requests that
	// Uber send email delivery notifications.
	// This field is optional and defaults to true.
	SendEmailNotifications bool `json:"send_email_notifications,omitempty"`

	// SendSMSNotifications if set requests that
	// Uber send SMS delivery notifications.
	// This field is optional and defaults to true.
	SendSMSNotifications bool `json:"send_sms_notifications,omitempty"`
}

type Phone struct {
	Number     string `json:"number"`
	SMSEnabled bool   `json:"sms_enabled"`
}

type CurrencyCode string

type DeliveryResponse struct {
	ID      string  `json:"delivery_id"`
	Fee     float32 `json:"fee"`
	QuoteID string  `json:"quote_id"`
	Status  Status  `json:"status"`

	Courier *Contact `json:"courier,omitempty"`

	OrderReferenceID string `json:"order_reference_id"`

	CurrencyCode CurrencyCode `json:"currency_code"`

	TrackingURL otils.NullableString `json:"tracking_url"`

	Items []*Item `json:"items"`

	Pickup  *Endpoint `json:"pickup"`
	Dropoff *Endpoint `json:"dropoff"`

	CreatedAt uint64 `json:"created_at"`

	// Batch is an optional object which
	// indicates whether a delivery should be
	// batched with other deliveries at pickup.
	Batch *Batch `json:"batch"`
}

type Batch struct {
	// Unique identifier of the batch. Deliveries
	// in the same batch share the same identifier.
	ID string `json:"batch_id"`

	// Count is the total number of deliveries in this batch.
	Count int64 `json:"count"`

	Deliveries []string `json:"deliveries,omitempty"`
}

var (
	errNilPickup  = errors.New("a non-nil pickup is required")
	errNilDropoff = errors.New("a non-nil pickup is required")

	errNilEndpointLocation = errors.New("a non-nil endpoint.location is required")
	errNilEndpointContact  = errors.New("a non-nil endpoint.contact is required")

	errInvalidItems = errors.New("expecting at least one valid item")
)

func (dr *DeliveryRequest) Validate() error {
	if dr == nil || dr.Pickup == nil {
		return errNilPickup
	}
	if err := dr.Pickup.Validate(); err != nil {
		return err
	}
	if err := dr.Dropoff.Validate(); err != nil {
		return err
	}
	if !atLeastOneValidItem(dr.Items) {
		return errInvalidItems
	}
	return nil
}

var (
	errInvalidQuantity = errors.New("quantity has to be > 0")
	errBlankItemTitle  = errors.New("item title has to be non-empty")
)

func (i *Item) Validate() error {
	if i == nil || i.Quantity <= 0 {
		return errInvalidQuantity
	}
	if i.Title == "" {
		return errBlankItemTitle
	}
	return nil
}

func atLeastOneValidItem(items []*Item) bool {
	for _, item := range items {
		if err := item.Validate(); err == nil {
			return true
		}
	}
	return false
}

func (e *Endpoint) Validate() error {
	if e == nil || e.Location == nil {
		return errNilEndpointLocation
	}
	if e.Contact == nil {
		return errNilEndpointContact
	}
	return nil
}

func (c *Client) RequestDelivery(req *DeliveryRequest) (*DeliveryResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	blob, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	theURL := fmt.Sprintf("%s/deliveries", c.baseURL())
	httpReq, err := http.NewRequest("POST", theURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}

	blob, _, err = c.doAuthAndHTTPReq(httpReq)
	if err != nil {
		return nil, err
	}
	dRes := new(DeliveryResponse)
	if err := json.Unmarshal(blob, dRes); err != nil {
		return nil, err
	}
	return dRes, nil
}
