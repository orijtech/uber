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

	"github.com/orijtech/otils"
)

type Receipt struct {
	// Unique identifier representing Request.
	RequestID string `json:"request_id"`

	// Subtotal = TotalFare - ChargeAdjustments
	Subtotal otils.NullableString `json:"subtotal"`

	// The fare after credits and refunds have been applied.
	TotalFare otils.NullableString `json:"total_fare"`

	// The total amount charged to the user's payment method.
	// This is the subtotal (split if applicable) with taxes included.
	TotalCharged otils.NullableString `json:"total_charged"`

	// The total amount still owed after attempting to charge the
	// user. May be null if amount was paid in full.
	TotalOwed otils.NullableFloat64 `json:"total_owed"`

	// The ISO 4217 currency code of the amounts.
	CurrencyCode otils.NullableString `json:"currency_code"`

	// Duration is the ISO 8601 HH:MM:SS
	// format of the time duration of the trip.
	Duration otils.NullableString `json:"currency_code"`

	// Distance of the trip charged.
	Distance otils.NullableString `json:"distance"`

	// UnitOfDistance is the localized unit of distance.
	UnitOfDistance otils.NullableString `json:"distance_label"`
}

var errEmptyReceiptID = errors.New("expecting a non-empty receiptID")

func (c *Client) RequestReceipt(receiptID string) (*Receipt, error) {
	if receiptID == "" {
		return nil, errEmptyReceiptID
	}

	fullURL := fmt.Sprintf("%s/requests/%s/receipt", c.baseURL(), receiptID)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	slurp, _, err := c.doAuthAndHTTPReq(req)
	if err != nil {
		return nil, err
	}

	receipt := new(Receipt)
	if err := json.Unmarshal(slurp, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}
