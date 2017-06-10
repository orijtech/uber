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

package uberhook_test

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"

	"github.com/orijtech/otils"
	"github.com/orijtech/uber/uberhook"
)

func Example_Server() {
	webhook, err := uberhook.New()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", webhook.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		event, err := uberhook.FparseEvent(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Got an event from Uber: %#v\n", event)
		fmt.Fprintf(w, "Successfully retrieved event!\n")
	})))

	go func() {
		nonHTTPSHandler := otils.RedirectAllTrafficTo("https://uberhook.example.com")
		if err := http.ListenAndServe(":80", nonHTTPSHandler); err != nil {
			log.Fatal(err)
		}
	}()

	domains := []string{
		"uberhook.example.com",
		"www.uberhook.example.com",
	}

	log.Fatal(http.Serve(autocert.NewListener(domains...), mux))
}
