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
	"net/url"
	"strings"
	"time"

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
	Location *Location `json:"location,omitempty"`
	Contact  *Contact  `json:"contact,omitempty"`

	// Special instructions for the endpoint. This field
	// is optional and it is limited to 256 characters.
	SpecialInstructions otils.NullableString `json:"special_instructions,omitempty"`

	SignatureRequired bool `json:"signature_required,omitempty"`

	// Indicates if the delivery includes alcohol. This
	// feature is only available to whitelisted businesses.
	IncludesAlcohol bool `json:"includes_alcohol,omitempty"`

	ETAMinutes int `json:"eta,omitempty"`

	TimestampUnix int64 `json:"timestamp,omitempty"`
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

type Delivery struct {
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

func (c *Client) RequestDelivery(req *DeliveryRequest) (*Delivery, error) {
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

	blob, _, err = c.doHTTPReq(httpReq)
	if err != nil {
		return nil, err
	}
	dRes := new(Delivery)
	if err := json.Unmarshal(blob, dRes); err != nil {
		return nil, err
	}
	return dRes, nil
}

var errBlankDeliveryID = errors.New("expecting a non-blank deliveryID")

// CancelDelivery cancels a delivery referenced by its ID. There are
// potential cancellation fees associated.
// See https://developer.uber.com/docs/deliveries/faq for more information.
func (c *Client) CancelDelivery(deliveryID string) error {
	deliveryID = strings.TrimSpace(deliveryID)
	if deliveryID == "" {
		return errBlankDeliveryID
	}
	theURL := fmt.Sprintf("%s/deliveries/%s/cancel", c.baseURL(), deliveryID)
	httpReq, err := http.NewRequest("POST", theURL, nil)
	if err != nil {
		return err
	}
	_, _, err = c.doHTTPReq(httpReq)
	return err
}

type DeliveryListRequest struct {
	Status        Status `json:"status,omitempty"`
	LimitPerPage  int64  `json:"limit"`
	MaxPageNumber int64  `json:"max_page,omitempty"`
	StartOffset   int64  `json:"offset"`

	ThrottleDurationMs int64 `json:"throttle_duration_ms"`
}

type DeliveryThread struct {
	Pages  chan *DeliveryPage `json:"-"`
	Cancel func()
}

type DeliveryPage struct {
	Err        error       `json:"error"`
	PageNumber int64       `json:"page_number,omitempty"`
	Deliveries []*Delivery `json:"deliveries,omitempty"`
}

type recvDelivery struct {
	Count             int64       `json:"count"`
	NextPageQuery     string      `json:"next_page"`
	PreviousPageQuery string      `json:"previous_page"`
	Deliveries        []*Delivery `json:"deliveries"`
}

type deliveryPager struct {
	Offset int64  `json:"offset"`
	Limit  int64  `json:"limit"`
	Status Status `json:"status"`
}

const (
	NoThrottle = -1

	defaultThrottleDurationMs = 150 * time.Millisecond
)

// ListDeliveries requires authorization with OAuth2.0 with
// the delivery scope set.
func (c *Client) ListDeliveries(dReq *DeliveryListRequest) (*DeliveryThread, error) {
	if dReq == nil {
		dReq = &DeliveryListRequest{Status: StatusReceiptReady}
	}

	baseURL := c.legacyV1BaseURL()
	fullURL := fmt.Sprintf("%s/deliveries", baseURL)
	qv, err := otils.ToURLValues(&deliveryPager{
		Limit:  dReq.LimitPerPage,
		Status: dReq.Status,
		Offset: dReq.StartOffset,
	})
	if err != nil {
		return nil, err
	}

	if len(qv) > 0 {
		fullURL = fmt.Sprintf("%s/deliveries?%s", baseURL, qv.Encode())
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	var errsList []string
	if want, got := parsedBaseURL.Scheme, parsedURL.Scheme; got != want {
		errsList = append(errsList, fmt.Sprintf("gotScheme=%q wantBaseScheme=%q", got, want))
	}
	if want, got := parsedBaseURL.Host, parsedURL.Host; got != want {
		errsList = append(errsList, fmt.Sprintf("gotHost=%q wantBaseHost=%q", got, want))
	}
	if len(errsList) > 0 {
		return nil, errors.New(strings.Join(errsList, "\n"))
	}

	maxPage := dReq.MaxPageNumber
	pageExceeded := func(pageNumber int64) bool {
		return maxPage > 0 && pageNumber >= maxPage
	}

	fullDeliveriesBaseURL := fmt.Sprintf("%s/deliveries", baseURL)
	resChan := make(chan *DeliveryPage)
	cancelChan, cancelFn := makeCancelParadigm()

	go func() {
		defer close(resChan)

		pageNumber := int64(0)
		throttleDurationMs := defaultThrottleDurationMs
		if dReq.ThrottleDurationMs == NoThrottle {
			throttleDurationMs = 0
		} else {
			throttleDurationMs = time.Duration(dReq.ThrottleDurationMs) * time.Millisecond
		}

		for {
			page := &DeliveryPage{PageNumber: pageNumber}

			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				page.Err = err
				resChan <- page
				return
			}

			slurp, _, err := c.doReq(req)
			if err != nil {
				page.Err = err
				resChan <- page
				return
			}

			recv := new(recvDelivery)
			if err := json.Unmarshal(slurp, recv); err != nil {
				page.Err = err
				resChan <- page
				return
			}

			page.Deliveries = recv.Deliveries
			resChan <- page
			pageNumber += 1
			pageToken := recv.NextPageQuery
			if pageExceeded(pageNumber) || pageToken == "" || len(recv.Deliveries) == 0 {
				return
			}

			fullURL = fmt.Sprintf("%s?%s", fullDeliveriesBaseURL, pageToken)

			select {
			case <-cancelChan:
				return
			case <-time.After(throttleDurationMs):
			}
		}
	}()

	return &DeliveryThread{Cancel: cancelFn, Pages: resChan}, nil
}
