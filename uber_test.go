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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/orijtech/uber"
)

func TestListPaymentMethods(t *testing.T) {
	client, err := uber.NewClient(testToken1)
	if err != nil {
		t.Fatalf("initializing client; %v", err)
	}

	testingRoundTripper := &tRoundTripper{route: listPaymentMethods}
	client.SetHTTPRoundTripper(testingRoundTripper)

	tests := [...]struct {
		wantPaylisting *uber.PaymentListing
		wantErr        bool
	}{
		0: {
			wantPaylisting: paymentListingFromFile("./testdata/list-payments-1.json"),
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
		wantBytes := jsonSerialize(pml)
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
	default:
		return makeResp("Not Found", http.StatusNotFound), nil
	}
}

var (
	respNoBearerTokenSet  = makeResp("Unauthorized: \"Bearer\" token missing", http.StatusUnauthorized)
	respUnauthorizedToken = makeResp("Unauthorized token", http.StatusUnauthorized)
)

func (trt *tRoundTripper) listPaymentMethodRoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != "GET" {
		return makeResp("only \"GET\" allowed", http.StatusMethodNotAllowed), nil
	}

	// Check the authorization next
	bearerTokenSplit := strings.Split(req.Header.Get("Authorization"), "Bearer")
	// Expecting a successful split to be of the form {"", " <The token>"}
	if len(bearerTokenSplit) < 2 {
		return respNoBearerTokenSet, nil
	}

	token := strings.TrimSpace(bearerTokenSplit[len(bearerTokenSplit)-1])
	if token == "" {
		return respNoBearerTokenSet, nil
	}

	if unauthorizedToken(token) {
		return respUnauthorizedToken, nil
	}

	resp := responseFromFileContent("./testdata/list-payments-1.json")
	return resp, nil
}

func responseFromFileContent(path string) *http.Response {
	f, err := os.Open(path)
	if err != nil {
		return makeResp(err.Error(), http.StatusInternalServerError)
	}

	prc, pwc := io.Pipe()
	go func() {
		defer pwc.Close()
		defer f.Close()
		io.Copy(pwc, f)
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
)

var authorizedTokens = map[string]bool{
	testToken1: true,
}

func unauthorizedToken(token string) bool {
	_, known := authorizedTokens[token]
	return !known
}

const (
	listPaymentMethods = "list-payment-methods"
)
