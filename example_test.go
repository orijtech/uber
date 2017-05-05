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
	"fmt"
	"log"

	"github.com/orijtech/uber"
)

func Example_client_ListPaymentMethods() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	listings, err := client.ListPaymentMethods()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("LastUsedD: %v\n", listings.LastUsedID)

	for i, method := range listings.Methods {
		fmt.Printf("#%d: ID: %q PaymentMethod: %q Description: %q\n",
			i, method.ID, method.PaymentMethod, method.Description)
	}
}

func Example_client_ListHistory() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	pagesChan, cancelChan, err := client.ListHistory(&uber.TripHistoryRequest{
		MaxPages:     4,
		LimitPerPage: 10,
		StartOffset:  0,
	})

	if err != nil {
		log.Fatal(err)
	}

	for page := range pagesChan {
		if page.Err != nil {
			fmt.Printf("Page: #%d err: %v\n", page.PageNumber, page.Err)
			continue
		}

		fmt.Printf("Page: #%d\n\n", page.PageNumber)
		for i, trip := range page.Trips {
			startCity := trip.StartCity
			if startCity.Name == "Tokyo" {
				fmt.Printf("aha found the first Tokyo trip, canceling any more requests!: %#v\n", trip)
				cancelChan <- true
				break
			}

			// Otherwise, continue listing
			fmt.Printf("Trip: #%d ==> %#v place: %#v\n", i, trip, startCity)
		}
	}
}

func Example_client_ListAllMyHistory() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	pagesChan, cancelChan, err := client.ListAllMyHistory()
	if err != nil {
		log.Fatal(err)
	}

	for page := range pagesChan {
		if page.Err != nil {
			fmt.Printf("Page: #%d err: %v\n", page.PageNumber, page.Err)
			continue
		}

		fmt.Printf("Page: #%d\n\n", page.PageNumber)
		for i, trip := range page.Trips {
			startCity := trip.StartCity
			if startCity.Name == "Edmonton" {
				fmt.Printf("aha found the trip from Edmonton, canceling the rest!: %#v\n", trip)
				cancelChan <- true
				break
			}

			// Otherwise, continue listing
			fmt.Printf("Trip: #%d ==> %#v place: %#v\n", i, trip, startCity)
		}
	}
}
