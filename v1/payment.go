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
	"strconv"

	"github.com/orijtech/otils"
)

type Payment struct {
	// ID is the unique identifier of the payment item.
	// If the payment is related to a trip, it is the same as TripID.
	ID       string `json:"payment_id,omitempty"`
	MethodID string `json:"payment_method_id,omitempty"`

	Category PaymentCategory `json:"category,omitempty"`

	Description   string        `json:"description,omitempty"`
	PaymentMethod PaymentMethod `json:"type,omitempty"`

	// DriverID is the unique identifier of the
	// driver who received or made the payment.
	DriverID string `json:"driver_id,omitempty"`

	// PartnerID is the unique identifier of the
	// Fleet Manager for the driver.
	PartnerID string `json:"partner_id,omitempty"`

	// TripID is the unique identifier of the trip associated
	// with the payment. It is only present for
	// PaymentCategory `Fare` otherwise it is null.
	TripID otils.NullableString `json:"trip_id,omitempty"`

	EventTime otils.NullableFloat64 `json:"event_time,omitempty"`

	// CashCollected is the amount collected in cash by the driver.
	// It is only set for Uber products that are enabled for cash payments.
	CashCollected otils.NullableFloat64 `json:"cash_collected,omitempty"`

	// Amount is the net payout to the driver. It is positive for
	// payments to the account, negative for charges to the account.
	Amount otils.NullableFloat64 `json:"amount,omitempty"`

	// CurrencyCode is the ISO 4217 currency code of the payment.
	CurrencyCode otils.NullableString `json:"currency_code,omitempty"`

	// Breakdown is the breakdown of the fare.
	Breakdown *FareBreakdown `json:"breakdown,omitempty"`

	// RiderFees details the fees paid to the driver by the rider
	// These fees are not reflected in the fare.
	RiderFees *ServiceFee `json:"rider_fees,omitempty"`
}

type PaymentMethod uint

const (
	PaymentUnknown PaymentMethod = iota

	// Last 2 digits of card e.g "***23" or the
	// obfuscated email address ("ga***@uber.com")
	// depending on the account identifier.
	PaymentAlipay

	// Last 2 digits of cards e.g "***23".
	PaymentApplePay
	PaymentAmericanExpress
	PaymentDiscover
	PaymentJCB
	PaymentLianLian
	PaymentMaestro
	PaymentMastercard
	PaymentPaypal
	PaymentPaytm
	PaymentUnionPay
	PaymentVisa

	// A descriptive name of the family account e.g "John Doe Family Shared".
	PaymentUberFamilyAccount

	// No description for these ones.
	PaymentAirtel
	PaymentAndroidPay
	PaymentCash
	PaymentUcharge
	PaymentZaakpay
)

var paymentMethodToString = map[PaymentMethod]string{
	PaymentAirtel:            "airtel",
	PaymentAlipay:            "alipay",
	PaymentApplePay:          "apple_pay",
	PaymentAmericanExpress:   "american_express",
	PaymentAndroidPay:        "android_pay",
	PaymentUberFamilyAccount: "family_account",
	PaymentCash:              "cash",
	PaymentDiscover:          "discover",
	PaymentJCB:               "jcb",
	PaymentLianLian:          "lianlian",
	PaymentMaestro:           "maestro",
	PaymentMastercard:        "mastercard",
	PaymentPaypal:            "paypal",
	PaymentPaytm:             "paytm",
	PaymentUcharge:           "ucharge",
	PaymentUnionPay:          "unionpay",
	PaymentUnknown:           "unknown",
	PaymentVisa:              "visa",
	PaymentZaakpay:           "zaakpay",
}

var stringToPaymentMethod map[string]PaymentMethod

func init() {
	stringToPaymentMethod = make(map[string]PaymentMethod)
	for paymentMethod, str := range paymentMethodToString {
		stringToPaymentMethod[str] = paymentMethod
	}
}

func (pm *PaymentMethod) PaymentMethodToString() string {
	if pm == nil {
		ppm := PaymentUnknown
		pm = &ppm
	}
	return paymentMethodToString[*pm]
}

func (pm PaymentMethod) String() string {
	return pm.PaymentMethodToString()
}

func StringToPaymentMethod(str string) PaymentMethod {
	pm, ok := stringToPaymentMethod[str]
	if !ok {
		pm = PaymentUnknown
	}
	return pm
}

var _ json.Unmarshaler = (*PaymentMethod)(nil)

func (pm *PaymentMethod) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	*pm = StringToPaymentMethod(unquoted)
	return nil
}

type PaymentListing struct {
	Methods []*Payment `json:"payment_methods,omitempty"`

	// The unique identifier of
	// the last used payment method.
	LastUsedID string `json:"last_used,omitempty"`
}

func (c *Client) ListPaymentMethods() (*PaymentListing, error) {
	fullURL := fmt.Sprintf("%s/payment-methods", c.baseURL())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en_US")

	slurp, _, err := c.doReq(req)
	if err != nil {
		return nil, err
	}

	listing := new(PaymentListing)
	if err := json.Unmarshal(slurp, listing); err != nil {
		return nil, err
	}
	return listing, nil
}
