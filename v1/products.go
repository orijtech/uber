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
	"reflect"
	"strings"

	"github.com/orijtech/otils"
)

type Product struct {
	UpfrontFareEnabled bool `json:"upfront_fare_enabled,omitempty"`
	// Capacity is the number of people that can be
	// accomodated by the product for example, 4 people.
	Capacity int `json:"capacity,omitempty"`

	// The unique identifier representing a specific
	// product for a given latitude and longitude.
	// For example, uberX in San Francisco will have
	// a different ID than uberX in Los Angeles.
	ID string `json:"product_id"`

	// PriceDetails details the basic price
	// (not including any surge pricing adjustments).
	// This field is nil for products with metered
	// fares(taxi) or upfront fares(uberPOOL).
	PriceDetails *PriceDetails `json:"price_details"`

	ImageURL    string `json:"image,omitempty"`
	CashEnabled bool   `json:"cash_enabled,omitempty"`
	Shared      bool   `json:"shared"`

	// An abbreviated description of the product.
	// It is localized according to `Accept-Language` header.
	ShortDescription string `json:"short_description"`

	DisplayName string `json:"display_name"`

	Description string `json:"description"`
}

type ProductGroup string

const (
	ProductRideShare ProductGroup = "rideshare"
	ProductUberX     ProductGroup = "uberx"
	ProductUberXL    ProductGroup = "uberxl"
	ProductUberBlack ProductGroup = "uberblack"
	ProductSUV       ProductGroup = "suv"
	ProductTaxi      ProductGroup = "taxi"
)

// ListProducts is a method that returns information about the
// Uber products offered at a given location.
// Some products such as uberEATS, are not returned by this
// endpoint, at least as of: Fri 23 Jun 2017 18:01:04 MDT.
// The results of this method do not reflect real-time availability
// of the products. Please use the EstimateTime method to determine
// real-time availability and ETAs of products.
// In some markets, the list of products returned from this endpoint
// may vary by the time of day due to time restrictions on
// when that product may be utilized.
func (c *Client) ListProducts(place *Place) ([]*Product, error) {
	qv, err := otils.ToURLValues(place)
	if err != nil {
		return nil, err
	}
	fullURL := fmt.Sprintf("%s/products?%s", c.baseURL(), qv.Encode())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	slurp, _, err := c.doReq(req)
	if err != nil {
		return nil, err
	}
	pWrap := new(productsWrap)
	if err := json.Unmarshal(slurp, pWrap); err != nil {
		return nil, err
	}
	return pWrap.Products, nil
}

var (
	errEmptyProductID = errors.New("expecting a non-empty productID")
	errBlankProduct   = errors.New("received a blank product back from the server")

	blankProductPtr = new(Product)
)

func (c *Client) ProductByID(productID string) (*Product, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, errEmptyProductID
	}
	fullURL := fmt.Sprintf("%s/products/%s", c.baseURL(), productID)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	slurp, _, err := c.doReq(req)
	if err != nil {
		return nil, err
	}
	product := new(Product)
	if err := json.Unmarshal(slurp, product); err != nil {
		return nil, err
	}
	if reflect.DeepEqual(product, blankProductPtr) {
		return nil, errBlankProduct
	}
	return product, nil
}

type productsWrap struct {
	Products []*Product `json:"products"`
}

type PriceDetails struct {
	// The base price of a trip.
	Base otils.NullableFloat64 `json:"base,omitempty"`

	// The minimum price of a trip.
	Minimum otils.NullableFloat64 `json:"minimum,omitempty"`

	// CostPerMinute is the charge per minute(if applicable for the product type).
	CostPerMinute otils.NullableFloat64 `json:"cost_per_minute,omitempty"`

	// CostPerDistanceUnit is the charge per
	// distance unit(if applicable for the product type).
	CostPerDistanceUnit otils.NullableFloat64 `json:"cost_per_distance,omitempty"`

	// DistanceUnit is the unit of distance used
	// to calculate the fare (either UnitMile or UnitKm)
	DistanceUnit Unit `json:"distance_unit,omitempty"`

	// Cancellation fee is what the rider has to pay after
	// they cancel the trip after the grace period.
	CancellationFee otils.NullableFloat64 `json:"cancellation_fee,omitempty"`

	CurrencyCode CurrencyCode `json:"currency_code,omitempty"`

	ServiceFees []*ServiceFee `json:"service_fees,omitempty"`
}

type ServiceFee struct {
	Name string `json:"name,omitempty"`

	Fee otils.NullableFloat64 `json:"fee,omitempty"`
}

type Unit otils.NullableString

const (
	UnitMile otils.NullableString = "mile"
	UnitKM   otils.NullableString = "km"
)
