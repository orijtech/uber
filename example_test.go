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
	"os"
	"time"

	"github.com/orijtech/uber/v1"
)

func Example_client_ListPaymentMethods() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
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
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
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
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
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
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
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
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
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
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
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
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	appliedPromoCode, err := client.ApplyPromoCode("uberd340ue")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("AppliedPromoCode: %#v\n", appliedPromoCode)
}

func Example_client_RequestReceipt() {
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

func Example_client_RetrieveHomeAddress() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	place, err := client.Place(uber.PlaceHome)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("My home address: %#v\n", place.Address)
}

func Example_client_RetrieveWorkAddress() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	place, err := client.Place(uber.PlaceWork)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("My work address: %#v\n", place.Address)
}

func Example_client_UpdateHomeAddress() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	updatedHome, err := client.UpdatePlace(&uber.PlaceParams{
		Place:   uber.PlaceHome,
		Address: "685 Market St, San Francisco, CA 94103, USA",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("My updated home address: %#v\n", updatedHome)
}

func Example_client_UpdateWorkAddress() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	updatedWork, err := client.UpdatePlace(&uber.PlaceParams{
		Place:   uber.PlaceWork,
		Address: "685 Market St, San Francisco, CA 94103, USA",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("My updated work address: %#v\n", updatedWork)
}

func Example_client_RequestMap() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	tripMapInfo, err := client.RequestMap("b5512127-a134-4bf4-b1ba-fe9f48f56d9d")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Visit the URL: %q for more information\n", tripMapInfo.URL)
}

func Example_client_OpenMap() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.OpenMapForTrip("64561dfe-87fa-41d7-807e-f364527b11cb"); err != nil {
		log.Fatal(err)
	}
}

func Example_client_UpfrontFare() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	upfrontFare, err := client.UpfrontFare(&uber.EstimateRequest{
		StartLatitude:  37.7752315,
		EndLatitude:    37.7752415,
		StartLongitude: -122.418075,
		EndLongitude:   -122.518075,
		SeatCount:      2,
	})
	if err != nil {
		log.Fatal(err)
	}

	if upfrontFare.SurgeInEffect() {
		fmt.Printf("Surge is in effect!\n")
		fmt.Printf("Please visit this URL to confirm %q then"+
			"request again", upfrontFare.Estimate.SurgeConfirmationURL)
		return
	}

	fmt.Printf("Fare: %#v\n", upfrontFare.Fare)
	fmt.Printf("Trip: %#v\n", upfrontFare.Trip)
}

func Example_client_RequestRide() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	ride, err := client.RequestRide(&uber.RideRequest{
		StartLatitude:  37.7752315,
		StartLongitude: -122.418075,
		EndLatitude:    37.7752415,
		EndLongitude:   -122.518075,
		PromptOnFare: func(fare *uber.UpfrontFare) error {
			if fare.Fare.Value >= 6.00 {
				return fmt.Errorf("exercise can't hurt instead of $6.00 for that walk!")
			}
			return nil
		},
	})
	if err != nil {
		log.Fatalf("ride request err: %v", err)
	}

	fmt.Printf("Your ride information: %+v\n", ride)
}

func Example_client_RequestDelivery() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	deliveryConfirmation, err := client.RequestDelivery(&uber.DeliveryRequest{
		Pickup: &uber.Endpoint{
			Contact: &uber.Contact{
				CompanyName:          "orijtech",
				Email:                "deliveries@orijtech.com",
				SendSMSNotifications: true,
			},
			Location: &uber.Location{
				PrimaryAddress: "Empire State Building",
				State:          "NY",
				Country:        "US",
			},
			SpecialInstructions: "Please ask guest services for \"I Man\"",
		},
		Dropoff: &uber.Endpoint{
			Contact: &uber.Contact{
				FirstName:   "delivery",
				LastName:    "bot",
				CompanyName: "Uber",

				SendEmailNotifications: true,
			},
			Location: &uber.Location{
				PrimaryAddress:   "530 W 113th Street",
				SecondaryAddress: "Floor 2",
				Country:          "US",
				PostalCode:       "10025",
				State:            "NY",
			},
		},
		Items: []*uber.Item{
			{
				Title:    "phone chargers",
				Quantity: 10,
			},
			{
				Title:    "Blue prints",
				Fragile:  true,
				Quantity: 1,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("The confirmation: %+v\n", deliveryConfirmation)
}

func Example_client_CancelDelivery() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = client.CancelDelivery("71a969ca-5359-4334-a7b7-5a1705869c51")
	if err == nil {
		log.Printf("Successfully canceled that delivery!")
	} else {
		log.Printf("Failed to cancel that delivery, err: %v", err)
	}
}

func Example_client_ListProducts() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	products, err := client.ListProducts(&uber.Place{
		Latitude:  38.8971,
		Longitude: -77.0366,
	})
	if err != nil {
		log.Fatal(err)
	}

	for i, product := range products {
		fmt.Printf("#%d: ID: %q Product: %#v\n", i, product.ID, product)
	}
}

func Example_client_ProductByID() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	product, err := client.ProductByID("bc300c14-c30d-4d3f-afcb-19b240c16a13")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The Product information: %#v\n", product)
}

func Example_client_ListDeliveries() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	delivRes, err := client.ListDeliveries(&uber.DeliveryListRequest{
		Status:      uber.StatusCompleted,
		StartOffset: 20,
	})
	if err != nil {
		log.Fatal(err)
	}

	itemCount := uint64(0)
	for page := range delivRes.Pages {
		if page.Err != nil {
			fmt.Printf("Page #%d err: %v", page.PageNumber, page.Err)
		}
		for i, delivery := range page.Deliveries {
			fmt.Printf("\t(%d): %#v\n", i, delivery)
			itemCount += 1
		}
		if itemCount >= 10 {
			delivRes.Cancel()
		}
	}
}

func Example_client_DriverProfile() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}
	prof, err := client.DriverProfile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Rating: %.2f\nFirst Name: %s\nLast Name: %s\n",
		prof.Rating, prof.FirstName, prof.LastName)
}

func Example_client_ListDriverPayments() {
	client, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}
	aWeekAgo := time.Now().Add(-1 * time.Hour * 7 * 24)
	yesterday := time.Now().Add(-1 * time.Hour * 24)
	payments, err := client.ListDriverPayments(&uber.DriverPaymentsQuery{
		StartDate: &aWeekAgo,
		EndDate:   &yesterday,
	})
	if err != nil {
		log.Fatal(err)
	}
	for page := range payments.Pages {
		if page.Err != nil {
			fmt.Printf("%d err: %v\n", page.PageNumber, page.Err)
			continue
		}
		for i, payment := range page.Payments {
			fmt.Printf("\t%d:: %#v\n", i, payment)
		}
		fmt.Println()
	}
}
