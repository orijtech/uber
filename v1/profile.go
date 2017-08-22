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

type Profile struct {
	// First name of the Uber user.
	FirstName string `json:"first_name,omitempty"`

	// Last name of the Uber user.
	LastName string `json:"last_name,omitempty"`

	// Email address of the Uber user.
	Email string `json:"email,omitempty"`

	// Image URL of the Uber user.
	PictureURL string `json:"picture,omitempty"`

	// Whether the user has confirmed their mobile number.
	MobileVerified bool `json:"mobile_verified"`

	// The promotion code for the user.
	// Can be used for rewards when referring
	// other users to Uber.
	PromoCode string `json:"promo_code,omitempty"`

	ID string `json:"uuid,omitempty"`

	Rating otils.NullableFloat64 `json:"rating,omitempty"`

	ActivationStatus ActivationStatus `json:"activation_status,omitempty"`

	DriverID string `json:"driver_id,omitempty"`
}

func (c *Client) RetrieveMyProfile() (*Profile, error) {
	return c.retrieveProfile("/me")
}

func (c *Client) retrieveProfile(path string) (*Profile, error) {
	fullURL := fmt.Sprintf("%s%s", c.baseURL(), path)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	slurp, _, err := c.doReq(req)
	if err != nil {
		return nil, err
	}
	prof := new(Profile)
	if err := json.Unmarshal(slurp, prof); err != nil {
		return nil, err
	}
	return prof, nil
}

type PromoCode struct {
	Description string `json:"description,omitempty"`
	Code        string `json:"promo_code,omitempty"`
}

var errNilPromoCode = errors.New("expecting a non-empty promoCode")

type PromoCodeRequest struct {
	CodeToApply string `json:"applied_promotion_codes"`
}

func (c *Client) ApplyPromoCode(promoCode string) (*PromoCode, error) {
	if promoCode == "" {
		return nil, errNilPromoCode
	}

	pcReq := &PromoCodeRequest{
		CodeToApply: promoCode,
	}

	blob, err := json.Marshal(pcReq)
	if err != nil {
		return nil, err
	}
	fullURL := fmt.Sprintf("%s/me", c.baseURL())
	req, err := http.NewRequest("PATCH", fullURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	slurp, _, err := c.doReq(req)
	if err != nil {
		return nil, err
	}

	appliedPromoCode := new(PromoCode)
	if err := json.Unmarshal(slurp, appliedPromoCode); err != nil {
		return nil, err
	}

	return appliedPromoCode, nil
}
