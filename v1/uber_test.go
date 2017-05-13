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

package uber_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	uberOAuth2 "github.com/orijtech/uber/oauth2"
	"github.com/orijtech/uber/v1"
)

func TestListPaymentMethods(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: listPaymentMethods}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		want    *uber.PaymentListing
		wantErr bool
	}{
		0: {
			want: paymentListingFromFile("./testdata/list-payments-1.json"),
		},
	}

	for i, tt := range tests {
		pml, err := client.ListPaymentMethods()
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: expected a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: got err: %v want nil error", i, err)
			continue
		}

		gotBytes := jsonSerialize(pml)
		wantBytes := jsonSerialize(tt.want)
		if !bytes.Equal(gotBytes, wantBytes) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBytes, wantBytes)
		}
	}
}

func TestListHistory(t *testing.T) {
	t.Skipf("Needs quite detailed data and intricate tests with paging")

	// client, err := uber.NewClient(testToken1)
	// if err != nil {
	// 	t.Fatalf("initializing client; %v", err)
	// }

	// if err != nil {
	// 	t.Fatalf("initializing client; %v", err)
	// }

	// testingRoundTripper := &tRoundTripper{route: listHistory}
	// client.SetHTTPRoundTripper(testingRoundTripper)

	// tests := [...]struct{}{}
}

func TestEstimatePrice(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: estimatePriceRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		ereq    *uber.EstimateRequest
		wantErr bool
		want    []*uber.PriceEstimate
	}{
		0: {
			ereq: &uber.EstimateRequest{
				StartLatitude:  37.7752315,
				EndLatitude:    37.7752415,
				StartLongitude: -122.418075,
				EndLongitude:   -122.518075,
			},
			want: priceEstimateFromFile("./testdata/estimate-1.json"),
		},
		1: {
			ereq:    nil,
			wantErr: true,
		},
	}

	for i, tt := range tests {
		estimatesChan, cancelPaging, err := client.EstimatePrice(tt.ereq)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d err: %v", i, err)
			continue
		}

		firstPage := <-estimatesChan
		// Then cancel it
		cancelPaging()

		if err := firstPage.Err; err != nil {
			t.Errorf("#%d paging err: %v, firstPage: %#v", i, err, firstPage)
			continue
		}
		estimates := firstPage.Estimates

		gotBlob, wantBlob := jsonSerialize(estimates), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func TestEstimateTime(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: estimateTimeRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		treq    *uber.EstimateRequest
		wantErr bool
		want    []*uber.TimeEstimate
	}{
		0: {
			treq: &uber.EstimateRequest{
				StartLatitude:  37.7752315,
				EndLatitude:    37.7752415,
				StartLongitude: -122.418075,
				EndLongitude:   -122.518075,
				ProductID:      "a1111c8c-c720-46c3-8534-2fcdd730040d",
			},
			want: timeEstimateFromFile("./testdata/time-estimate-1.json"),
		},
		1: {
			treq:    nil,
			wantErr: true,
		},
	}

	for i, tt := range tests {
		estimatesChan, cancelPaging, err := client.EstimateTime(tt.treq)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d err: %v", i, err)
			continue
		}

		firstPage := <-estimatesChan
		// Then cancel it
		cancelPaging()

		if err := firstPage.Err; err != nil {
			t.Errorf("#%d paging err: %v, firstPage: %#v", i, err, firstPage)
			continue
		}
		estimates := firstPage.Estimates

		gotBlob, wantBlob := jsonSerialize(estimates), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func TestRetrieveMyProfile(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: retrieveProfileRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	invalidToken := fmt.Sprintf("%v", time.Now().Unix())

	tests := [...]struct {
		wantErr     bool
		bearerToken string
		want        *uber.Profile
	}{
		0: {
			bearerToken: testToken1,
			want:        profileFromFileByToken(testToken1),
		},
		1: {
			bearerToken: invalidToken,
			wantErr:     true,
		},
	}

	for i, tt := range tests {
		client.SetBearerToken(tt.bearerToken)
		prof, err := client.RetrieveMyProfile()
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		gotBlob, wantBlob := jsonSerialize(prof), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func TestApplyPromoCode(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: applyPromoCodeRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		wantErr   bool
		promoCode string
		want      *uber.PromoCode
	}{
		0: {
			promoCode: promoCode1,
			want:      promoCodeFromFileByToken(promoCode1),
		},
		1: {
			// Try with a random promo code that's unauthorized.
			promoCode: fmt.Sprintf("%v", time.Now().Unix()),
			wantErr:   true,
		},
	}

	for i, tt := range tests {
		appliedPromoCode, err := client.ApplyPromoCode(tt.promoCode)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		gotBlob, wantBlob := jsonSerialize(appliedPromoCode), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

const (
	requestID1 = "b5512127-a134-4bf4-b1ba-fe9f48f56d9d"
)

func TestPlaceRetrieval(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: getPlacesRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		wantErr bool
		place   uber.PlaceName
		want    *uber.Place
	}{
		0: {
			place: "home",
			want:  placeFromFile("685-market"),
		},
		1: {
			place: "work",
			want:  placeFromFile("wallaby-way"),
		},
		2: {
			place:   "workz",
			wantErr: true,
		},
		3: {
			place:   "",
			wantErr: true,
		},
	}

	for i, tt := range tests {
		place, err := client.Place(tt.place)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		gotBlob, wantBlob := jsonSerialize(place), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

var testOAuth2Token1 = &oauth2.Token{
	AccessToken:  testOAuth2AccessToken1,
	TokenType:    "Bearer",
	RefreshToken: "uber-test-refresh-token",
}

func TestUpfrontFare(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}
	testingRoundTripper := &tRoundTripper{route: upfrontFareRoute}
	transport := uberOAuth2.TransportWithBase(testOAuth2Token1, testingRoundTripper)
	client.SetHTTPRoundTripper(transport)

	tests := [...]struct {
		wantErr bool
		req     *uber.EstimateRequest
		want    *uber.UpfrontFare
	}{
		0: {
			req: &uber.EstimateRequest{
				StartPlace: uber.PlaceHome,

				EndLatitude:  37.7752415,
				EndLongitude: -122.518075,
			},
			want: upfrontFareFromFileByID("surge"),
		},
		1: {
			req: &uber.EstimateRequest{
				StartLatitude:  37.7752415,
				StartLongitude: -122.518075,

				EndPlace: uber.PlaceWork,
			},
			want: upfrontFareFromFileByID("no-surge"),
		},
		2: {
			req:     nil,
			wantErr: true,
		},
	}

	for i, tt := range tests {
		upfrontFare, err := client.UpfrontFare(tt.req)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		if tt.want == nil {
			t.Errorf("#%d: inconsistency want is nil", i)
			continue
		}

		gotBlob, wantBlob := jsonSerialize(upfrontFare), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func TestPlaceUpdate(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: updatePlacesRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		wantErr bool
		params  *uber.PlaceParams
		want    *uber.Place
	}{
		0: {
			params: &uber.PlaceParams{Place: uber.PlaceHome, Address: "P Sherman 42 Wallaby Way Sydney"},
			want:   placeFromFile("wallaby-way"),
		},
		1: {
			params: &uber.PlaceParams{Place: uber.PlaceWork, Address: "685 Market St, San Francisco, CA 94103, USA"},
			want:   placeFromFile("685-market"),
		},
		2: {
			params:  &uber.PlaceParams{},
			wantErr: true,
		},

		3: {
			// No place was specified.
			params:  &uber.PlaceParams{Address: "685 Market St, San Francisco, CA 94103, USA"},
			wantErr: true,
		},

		4: {
			// No address was specified.
			params:  &uber.PlaceParams{Place: uber.PlaceHome},
			wantErr: true,
		},

		5: {
			// No address was specified.
			params:  &uber.PlaceParams{Place: uber.PlaceWork},
			wantErr: true,
		},
	}

	for i, tt := range tests {
		place, err := client.UpdatePlace(tt.params)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		gotBlob, wantBlob := jsonSerialize(place), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func mapFromFile(tripID string) *uber.Map {
	diskPath := mapPathFromRequestID(tripID)
	save := new(uber.Map)
	if err := readFromFileAndDeserialize(diskPath, save); err != nil {
		return nil
	}
	return save
}

func TestRequestMap(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: getMapRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		wantErr   bool
		requestID string
		want      *uber.Map
	}{
		0: {
			requestID: requestID1,
			want:      mapFromFile(requestID1),
		},
		1: {
			// Try with a random requestID.
			requestID: fmt.Sprintf("%v", time.Now().Unix()),
			wantErr:   true,
		},
	}

	for i, tt := range tests {
		mapInfo, err := client.RequestMap(tt.requestID)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		gotBlob, wantBlob := jsonSerialize(mapInfo), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func TestRequestReceipt(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: requestReceiptRoute}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		wantErr   bool
		requestID string
		want      *uber.Receipt
	}{
		0: {
			requestID: requestID1,
			want:      receiptFromFile(requestID1),
		},
		1: {
			// Try with a random requestID.
			requestID: fmt.Sprintf("%v", time.Now().Unix()),
			wantErr:   true,
		},
	}

	for i, tt := range tests {
		receipt, err := client.RequestReceipt(tt.requestID)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		gotBlob, wantBlob := jsonSerialize(receipt), jsonSerialize(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func profileTokenPath(tokenSuffix string) string {
	return fmt.Sprintf("./testdata/profile-%s.json", tokenSuffix)
}

func promoCodePath(suffix string) string {
	return fmt.Sprintf("./testdata/promo-code-%s.json", suffix)
}

func promoCodeFromFileByToken(promoCodeSuffix string) *uber.PromoCode {
	path := promoCodePath(promoCodeSuffix)
	promoCode := new(uber.PromoCode)
	if err := readFromFileAndDeserialize(path, promoCode); err != nil {
		return nil
	}
	return promoCode
}

func profileFromFileByToken(tokenSuffix string) *uber.Profile {
	path := profileTokenPath(tokenSuffix)
	prof := new(uber.Profile)
	if err := readFromFileAndDeserialize(path, prof); err != nil {
		return nil
	}
	return prof
}

func jsonSerialize(v interface{}) []byte {
	blob, _ := json.Marshal(v)
	return blob
}

type tRoundTripper struct {
	route string
}

func makeResp(status string, code int) *http.Response {
	res := &http.Response{
		StatusCode: code, Status: status,
		Header: make(http.Header),
		Body:   http.NoBody,
	}

	return res
}

var _ http.RoundTripper = (*tRoundTripper)(nil)

func (trt *tRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	switch trt.route {
	case listPaymentMethods:
		return trt.listPaymentMethodRoundTrip(req)
	case estimatePriceRoute:
		return trt.estimatePriceRoundTrip(req)
	case estimateTimeRoute:
		return trt.estimateTimeRoundTrip(req)
	case retrieveProfileRoute:
		return trt.retrieveProfileRoundTrip(req)
	case applyPromoCodeRoute:
		return trt.applyPromoCodeRoundTrip(req)
	case requestReceiptRoute:
		return trt.requestReceiptRoundTrip(req)
	case getMapRoute:
		return trt.requestMapRoundTrip(req)
	case getPlacesRoute:
		return trt.getPlacesRoundTrip(req)
	case updatePlacesRoute:
		return trt.updatePlacesRoundTrip(req)
	case upfrontFareRoute:
		return trt.upfrontFareRoundTrip(req)
	default:
		return makeResp("Not Found", http.StatusNotFound), nil
	}
}

var (
	respNoBearerTokenSet  = makeResp("Unauthorized: \"Bearer\" token missing", http.StatusUnauthorized)
	respUnauthorizedToken = makeResp("Unauthorized token", http.StatusUnauthorized)
)

func prescreenAuthAndMethod(req *http.Request, wantMethod string) (*http.Response, string, error) {
	if req.Method != wantMethod {
		msg := fmt.Sprintf("only %q allowed not %q", wantMethod, req.Method)
		return makeResp(msg, http.StatusMethodNotAllowed), "", nil
	}

	// Check the authorization next
	bearerTokenSplit := strings.Split(req.Header.Get("Authorization"), "Bearer")
	// Expecting a successful split to be of the form {"", " <The token>"}
	if len(bearerTokenSplit) < 2 {
		return respNoBearerTokenSet, "", nil
	}

	token := strings.TrimSpace(bearerTokenSplit[len(bearerTokenSplit)-1])
	if token == "" {
		return respNoBearerTokenSet, "", nil
	}

	if unauthorizedToken(token) {
		return respUnauthorizedToken, "", nil
	}

	// All passed nothing to report back.
	return nil, token, nil
}

func (trt *tRoundTripper) applyPromoCodeRoundTrip(req *http.Request) (*http.Response, error) {
	badAuthResp, _, err := prescreenAuthAndMethod(req, "PATCH")
	if badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	slurp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return makeResp(err.Error(), http.StatusInternalServerError), nil
	}

	preq := new(uber.PromoCodeRequest)
	if err := json.Unmarshal(slurp, preq); err != nil {
		return makeResp(err.Error(), http.StatusInternalServerError), nil
	}

	resp := responseFromFileContent(promoCodePath(preq.CodeToApply))
	return resp, nil

}

func (trt *tRoundTripper) retrieveProfileRoundTrip(req *http.Request) (*http.Response, error) {
	badAuthResp, token, err := prescreenAuthAndMethod(req, "GET")
	if badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	resp := responseFromFileContent(profileTokenPath(token))
	return resp, nil
}

func (trt *tRoundTripper) estimateTimeRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "GET"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	resp := responseFromFileContent("./testdata/time-estimate-1.json")
	return resp, nil
}

func (trt *tRoundTripper) estimatePriceRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "GET"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	resp := responseFromFileContent("./testdata/price-estimate-1.json")
	return resp, nil
}

var addressesToIDs = map[string]string{
	"home": "685-market",
	"work": "wallaby-way",

	"P Sherman 42 Wallaby Way Sydney":             "wallaby-way",
	"685 Market St, San Francisco, CA 94103, USA": "685-market",
}

func (trt *tRoundTripper) getPlacesRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "GET"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	splits := strings.Split(req.URL.Path, "/")
	if len(splits) < 2 {
		resp := makeResp("expecting the place", http.StatusBadRequest)
		return resp, nil
	}

	placeID := splits[len(splits)-1]
	switch uber.PlaceName(placeID) {
	case uber.PlaceHome, uber.PlaceWork:
	default:
		return makeResp("unknown place", http.StatusBadRequest), nil
	}

	pathID := addressesToIDs[placeID]
	diskPath := placePathFromID(pathID)
	return responseFromFileContent(diskPath), nil
}

func (trt *tRoundTripper) upfrontFareRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "POST"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	defer req.Body.Close()

	slurp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return makeResp(err.Error(), http.StatusBadRequest), nil
	}

	if len(slurp) < 20 { // Expecting at least body
		return makeResp("expecting a body", http.StatusBadRequest), nil
	}

	esReq := new(uber.EstimateRequest)
	if err := json.Unmarshal(slurp, esReq); err != nil {
		return makeResp(err.Error(), http.StatusBadRequest), nil
	}

	diskPath := fareEstimatePath(surgeIDFromPlace(esReq.EndPlace))
	return responseFromFileContent(diskPath), nil
}

func fareEstimatePath(suffix string) string {
	return fmt.Sprintf("./testdata/fare-estimate-%s.json", suffix)
}

func upfrontFareFromFileByID(id string) *uber.UpfrontFare {
	path := fareEstimatePath(id)
	save := new(uber.UpfrontFare)
	if err := readFromFileAndDeserialize(path, save); err != nil {
		return nil
	}
	return save

}

func surgeIDFromPlace(place uber.PlaceName) string {
	switch place {
	default:
		return "surge"
	case uber.PlaceWork:
		return "no-surge"
	}
}

func (trt *tRoundTripper) updatePlacesRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "PUT"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	defer req.Body.Close()

	slurp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return makeResp(err.Error(), http.StatusBadRequest), nil
	}

	pp := new(uber.PlaceParams)
	if err := json.Unmarshal(slurp, pp); err != nil {
		return makeResp(err.Error(), http.StatusBadRequest), nil
	}
	address := strings.TrimSpace(pp.Address)
	if address == "" {
		return makeResp("expecting a non-empty address", http.StatusBadRequest), nil
	}

	pathID := addressesToIDs[address]
	diskPath := placePathFromID(pathID)
	return responseFromFileContent(diskPath), nil
}

func mapPathFromRequestID(tripID string) string {
	return fmt.Sprintf("./testdata/map-%s.json", tripID)
}

func (trt *tRoundTripper) requestMapRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "GET"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}

	pathSplits := strings.Split(req.URL.Path, "/")
	if len(pathSplits) < 2 {
		resp := makeResp("expecting the requestID", http.StatusBadRequest)
		return resp, nil
	}

	// second last item
	requestID := pathSplits[len(pathSplits)-2]
	diskPath := mapPathFromRequestID(requestID)
	resp := responseFromFileContent(diskPath)
	return resp, nil
}

func (trt *tRoundTripper) requestReceiptRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "GET"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}

	pathSplits := strings.Split(req.URL.Path, "/")
	if len(pathSplits) < 2 {
		resp := makeResp("expecting the requestID", http.StatusBadRequest)
		return resp, nil
	}

	// second last item
	requestID := pathSplits[len(pathSplits)-2]
	diskPath := receiptPathFromRequestID(requestID)
	resp := responseFromFileContent(diskPath)
	return resp, nil
}

func (trt *tRoundTripper) listPaymentMethodRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, _, err := prescreenAuthAndMethod(req, "GET"); badAuthResp != nil || err != nil {
		return badAuthResp, err
	}
	resp := responseFromFileContent("./testdata/list-payments-1.json")
	return resp, nil
}

func responseFromFileContent(path string) *http.Response {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return makeResp(err.Error(), http.StatusInternalServerError)
	}

	prc, pwc := io.Pipe()
	go func() {
		defer pwc.Close()
		pwc.Write(data)
	}()

	resp := makeResp("200 OK", http.StatusOK)
	resp.Body = prc
	return resp
}

func receiptPathFromRequestID(requestID string) string {
	return fmt.Sprintf("./testdata/receipt-%s.json", requestID)
}

func paymentListingFromFile(path string) *uber.PaymentListing {
	save := new(uber.PaymentListing)
	if err := readFromFileAndDeserialize(path, save); err != nil {
		return nil
	}
	return save
}

func timeEstimateFromFile(path string) []*uber.TimeEstimate {
	save := new(uber.TimeEstimatesPage)
	if err := readFromFileAndDeserialize(path, save); err != nil {
		return nil
	}
	return save.Estimates
}

func priceEstimateFromFile(path string) []*uber.PriceEstimate {
	save := new(uber.PriceEstimatesPage)
	if err := readFromFileAndDeserialize(path, save); err != nil {
		return nil
	}
	return save.Estimates
}

func receiptFromFile(requestID string) *uber.Receipt {
	save := new(uber.Receipt)
	path := receiptPathFromRequestID(requestID)
	if err := readFromFileAndDeserialize(path, save); err != nil {
		return nil
	}
	return save
}

func placePathFromID(placeID string) string {
	return fmt.Sprintf("./testdata/place-%s.json", placeID)
}

func placeFromFile(placeID string) *uber.Place {
	save := new(uber.Place)
	path := placePathFromID(placeID)
	if err := readFromFileAndDeserialize(path, save); err != nil {
		return nil
	}
	return save
}

func readFromFileAndDeserialize(path string, save interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	slurp, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(slurp, save)
}

const (
	testToken1 = "TEST_TOKEN-1"

	testOAuth2AccessToken1 = "uber-test-access-token"

	promoCode1 = "pc1"
)

var authorizedTokens = map[string]bool{
	testToken1: true,

	testOAuth2AccessToken1: true,
}

func unauthorizedToken(token string) bool {
	_, known := authorizedTokens[token]
	return !known
}

const (
	listPaymentMethods   = "list-payment-methods"
	estimatePriceRoute   = "estimate-prices"
	estimateTimeRoute    = "estimate-times"
	retrieveProfileRoute = "retrieve-profile"
	applyPromoCodeRoute  = "apply-promo-code"
	getMapRoute          = "get-map"
	requestReceiptRoute  = "request-receipt"
	getPlacesRoute       = "get-places"
	updatePlacesRoute    = "update-places"
	upfrontFareRoute     = "upfront-fare"
)
