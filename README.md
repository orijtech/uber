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
```
