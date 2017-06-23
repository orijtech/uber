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
	"strings"
	"sync"
)

func makeCancelParadigm() (<-chan bool, func()) {
	var cancelOnce sync.Once
	cancelChan := make(chan bool, 1)
	cancelFn := func() {
		cancelOnce.Do(func() {
			close(cancelChan)
		})
	}

	return cancelChan, cancelFn
}

type Error struct {
	Meta   interface{}         `json:"meta"`
	Errors []*statusCodedError `json:"errors"`

	memoized string
}

func (ue *Error) Error() string {
	if ue == nil {
		return ""
	}
	if ue.memoized != "" {
		return ue.memoized
	}

	// Otherwise create it
	var errsList []string
	for _, sce := range ue.Errors {
		errsList = append(errsList, sce.Error())
	}
	ue.memoized = strings.Join(errsList, "\n")
	return ue.memoized
}

var _ error = (*Error)(nil)
var _ error = (*statusCodedError)(nil)

type statusCodedError struct {
	// The json tags are intentionally reversed
	// because an uber status coded error looks
	// like this:
	// {
	//    "status":404,
	//    "code":"unknown_place_id",
	//    "title":"Could not resolve the given place_id."
	// }
	// of which the above definitions seem reversed compared to
	// Go's net/http Request where Status is a message and StatusCode is an int.
	Code    int    `json:"status"`
	Message string `json:"code"`
	Title   string `json:"title"`

	memoizedErr string
}

func (sce *statusCodedError) Error() string {
	if sce == nil {
		return ""
	}
	if sce.memoizedErr == "" {
		blob, _ := json.Marshal(sce)
		sce.memoizedErr = string(blob)
	}
	return sce.memoizedErr
}

// GeoCoord is a struct to simplify unpacking
// the results of GeoJSON sent in the form
//    [Lat, Lon]
// which becomes inconvenient when used in
// code ie g[0], g[1] instead of just
// g.Latitude, g.Longitude.
type GeoCoord struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

var _ json.Unmarshaler = (*GeoCoord)(nil)
var _ json.Marshaler = (*GeoCoord)(nil)

func (gc *GeoCoord) UnmarshalJSON(b []byte) error {
	var coordSlice []float32
	if err := json.Unmarshal(b, &coordSlice); err != nil {
		return err
	}

	gc.Latitude = coordSlice[0]
	gc.Longitude = coordSlice[1]
	return nil
}

func (gc *GeoCoord) MarshalJSON() ([]byte, error) {
	outRepr := []float32{gc.Latitude, gc.Longitude}
	return json.Marshal(outRepr)
}

type FeatureType string

type Feature struct {
	Type       FeatureType `json:"type"`
	Properties interface{} `json:"properties"`
	Geometry   *Geometry   `json:"geometry"`
}

type Geometry struct {
	Type string `json:"type"`

	Coordinates []*GeoCoord `json:"coordinates"`
}
