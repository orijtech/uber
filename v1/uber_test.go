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
		estimatesChan, cancelChan, err := client.EstimatePrice(tt.ereq)
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
		cancelChan <- true

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
		estimatesChan, cancelChan, err := client.EstimateTime(tt.treq)
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
		cancelChan <- true

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
	authResp, _, err := prescreenAuthAndMethod(req, "PATCH")
	if authResp != nil || err != nil {
		return authResp, err
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
	authResp, token, err := prescreenAuthAndMethod(req, "GET")
	if authResp != nil || err != nil {
		return authResp, err
	}
	resp := responseFromFileContent(profileTokenPath(token))
	return resp, nil
}

func (trt *tRoundTripper) estimateTimeRoundTrip(req *http.Request) (*http.Response, error) {
	if authResp, _, err := prescreenAuthAndMethod(req, "GET"); authResp != nil || err != nil {
		return authResp, err
	}
	resp := responseFromFileContent("./testdata/time-estimate-1.json")
	return resp, nil
}

func (trt *tRoundTripper) estimatePriceRoundTrip(req *http.Request) (*http.Response, error) {
	if authResp, _, err := prescreenAuthAndMethod(req, "GET"); authResp != nil || err != nil {
		return authResp, err
	}
	resp := responseFromFileContent("./testdata/price-estimate-1.json")
	return resp, nil
}

func (trt *tRoundTripper) listPaymentMethodRoundTrip(req *http.Request) (*http.Response, error) {
	if authResp, _, err := prescreenAuthAndMethod(req, "GET"); authResp != nil || err != nil {
		return authResp, err
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

	promoCode1 = "pc1"
)

var authorizedTokens = map[string]bool{
	testToken1: true,
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
)
