# uber
Uber API client in Go

* Requirement:
To use client v1, you'll need to set
+ `UBER_TOKEN_KEY`

Sample usage: You can see file 
[example_test.go](./example_test.go)

* Preamble:
```go
import (
	"fmt"
	"log"

	"github.com/orijtech/uber/v1"
)
```

* List my payment methods
```go
func allMyPayments() {
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
```

* List all my history
```go
func searchingForFirstEdmontonTrip() {
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
```

* Use a promo code for your account
```go
func applyPromoCode() {
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
```

* Retrieve your profile
```go
func retrieveMyProfile() {
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
```

* Get price estimates
```go
func getPriceEstimates() {
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
```

* Get time estimates
```go
func getTimeEstimates() {
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
```

* Retrieve a receipt
```go
func retrieveReceipt() {
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
```

* Retrieve your home address
```go
func retrieveMyHomeAddress() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	place, err := client.Place(uber.AddressHome)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("My home address: %#v\n", place.Address)
}
```

* Retrieve your work address
```go
func retrieveMyWorkAddress() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	place, err := client.Place(uber.AddressWork)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("My work address: %#v\n", place.Address)
}
```

* Update your home address
```go
func updateMyHomeAddress() {
	client, err := uber.NewClient()
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
```

* Update your work address
```go
func updateMyWorkAddress() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	updatedWork, err := client.UpdatePlace(&uber.PlaceParams{
		Place:   uber.PlaceWork,
		Address: "685 Market St, San Francisco, CA 94103, USA",
	})
	if err != nil {
		log.Fatalf("work failed; %v", err)
	}

	fmt.Printf("My updated work address: %#v\n", updatedWork)
}
```

* Retrieve the map for a trip
```go
func requestMap() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	tripMapInfo, err := client.RequestMap("b5512127-a134-4bf4-b1ba-fe9f48f56d9d")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Visit the URL: %q for more information\n", tripMapInfo.URL)
}
```

* Open the map for a trip in your web browser
```go
func openTheTripInBrowser() {
	client, err := uber.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := client.OpenMapForTrip("b5512127-a134-4bf4-b1ba-fe9f48f56d9d"); err != nil {
		log.Fatal(err)
	}
}
```
