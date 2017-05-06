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

	"github.com/orijtech/uber/v1"
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

	pagesChan, cancelPaging, err := client.ListHistory(&uber.Pager{
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
				cancelPaging()
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

	pagesChan, cancelPaging, err := client.ListAllMyHistory()
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
				cancelPaging()
				break
			}

			// Otherwise, continue listing
			fmt.Printf("Trip: #%d ==> %#v place: %#v\n", i, trip, startCity)
		}
	}
}

func Example_client_EstimatePrice() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	estimatesPageChan, cancelPaging, err := client.EstimatePrice(&uber.EstimateRequest{
		StartLatitude:  37.7752315,
		EndLatitude:    37.7752415,
		StartLongitude: -122.418075,
		EndLongitude:   -122.518075,
		SeatCount:      2,
	})

	if err != nil {
		log.Fatal(err)
	}

	itemCount := uint64(0)
	for page := range estimatesPageChan {
		if page.Err != nil {
			fmt.Printf("PageNumber: #%d err: %v", page.PageNumber, page.Err)
			continue
		}

		for i, estimate := range page.Estimates {
			itemCount += 1
			fmt.Printf("Estimate: #%d ==> %#v\n", i, estimate)
		}

		if itemCount >= 23 {
			cancelPaging()
		}
	}
}

func Example_client_EstimateTime() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	estimatesPageChan, cancelPaging, err := client.EstimateTime(&uber.EstimateRequest{
		StartLatitude:  37.7752315,
		EndLatitude:    37.7752415,
		StartLongitude: -122.418075,
		EndLongitude:   -122.518075,

		// Comment out to search only for estimates for: uberXL
		// ProductID: "821415d8-3bd5-4e27-9604-194e4359a449",
	})

	if err != nil {
		log.Fatal(err)
	}

	itemCount := uint64(0)
	for page := range estimatesPageChan {
		if page.Err != nil {
			fmt.Printf("PageNumber: #%d err: %v", page.PageNumber, page.Err)
			continue
		}

		for i, estimate := range page.Estimates {
			itemCount += 1
			fmt.Printf("Estimate: #%d ==> %#v\n", i, estimate)
		}

		if itemCount >= 23 {
			cancelPaging()
		}
	}
}

func Example_client_RetrieveMyProfile() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	myProfile, err := client.RetrieveMyProfile()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Here is my profile: %#v\n", myProfile)
}

func Example_client_ApplyPromoCode() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	appliedPromoCode, err := client.ApplyPromoCode("uberd340ue")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("AppliedPromoCode: %#v\n", appliedPromoCode)
}

func ExampleRequestReceipt() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	receipt, err := client.RequestReceipt("b5512127-a134-4bf4-b1ba-fe9f48f56d9d")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("That receipt: %#v\n", receipt)
}
