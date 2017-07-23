# uber
Uber API client in Go

## Table of contents
- [Requirements](#requirements)
- [CLI](#cli)
  - [Installation](#installation)
  - [init](#init)
  - [history](#history)
  - [order](#order)
  - [payments](#payments)
- [SDK Usage](#sdk-usage)

## Requirements:
To use client v1, you'll need to set
+ `UBER_TOKEN_KEY`

## SDK usage
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

* Request a ride:
```go
func requestARide() {
	client, err := uber.NewClientFromOAuth2File("./testdata/.uber/credentials.json")
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

* Request a delivery
```go
func requestDelivery() {
	client, err := uber.NewClientFromOAuth2File("./testdata/.uber/credentials.json")
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
```

* Cancel a delivery
```go
func cancelDelivery() {
	client, err := uber.NewClientFromOAuth2File("./testdata/.uber/credentials.json")
	if err != nil {
		log.Fatal(err)
	}

	err := client.CancelDelivery("71a969ca-5359-4334-a7b7-5a1705869c51")
	if err == nil {
		log.Printf("Successfully canceled that delivery!")
	} else {
		log.Printf("Failed to cancel that delivery, err: %v", err)
	}
}
```

## CLI
### Installation
```go
$ go get -u -v github.com/orijtech/uber/cmd/uber
```

### init
init initializes the context and authorization for your Uber app in the current working directory

```shell
$ go get -u -v github.com/orijtech/uber/cmd/uber
$ uber init
Please visit this URL for the auth dialog: https://login.uber.com/oauth/v2/authorize?access_type=offline&client_id=a_client_id&redirect_uri=https%3A%2F%2Fexample.org/uber&response_type=code&scope=profile+request+history+places+request_receipt+delivery&state=15004223370.604660
```
which after successful authorization will give you a notice in your browser, to return to
your terminal and will save the token to a file on disk, for example:
```shell
Successfully saved your OAuth2.0 token to "/Users/orijtech/uber-account/.uber/credentials.json"
```

From then on, for that Uber account, please go into that directory "/Users/orijtech/uber-account/"
in order to use that account

### history
history allows you to retrieve and examine your previous trips in a tabular form

```shell
$ uber history -h
```
for all available options.

* List your last 3 trips
```shell
$ uber history --limit-per-page 3 --max-page 1

Page: #1
+--------+---------------+-------------------------+----------+-------+--------------------------------------+
| TRIP # |  CITY         |          DATE           | DURATION | MILES |              REQUESTID               |
+--------+---------------+-------------------------+----------+-------+--------------------------------------+
|      1 | Denver        | 2017/07/15 21:47:44 MDT | 7m31s    | 3.211 | 8e7f479c-63e2-4ccc-babd-8671771485c3 |
+--------+---------------+-------------------------+----------+-------+--------------------------------------+
|      2 | San Francisco | 2017/07/13 18:11:06 MDT | 14m16s   | 3.694 | d521aed9-e9bc-4673-9109-25d9ce5c434c |
+--------+---------------+-------------------------+----------+-------+--------------------------------------+
|      3 | London        | 2017/06/25 16:17:43 MDT | 13m35s   | 3.318 | 1ce3cccb-2e09-4920-ad80-d00a4645f9ce |
+--------+---------------+-------------------------+----------+-------+--------------------------------------+
```

### order
order allows you to order an Uber to any location and destination
```shell
$ uber order
Start Point: Redwood City Cinemark
+--------+--------------------------------+-----------+-----------+-------------+
| CHOICE |              NAME              | RELEVANCE | LATITUDE  |  LONGITUDE  |
+--------+--------------------------------+-----------+-----------+-------------+
|      0 | Cinemark 20 Redwood City,      | 98.70%    | 37.485912 | -122.228752 |
|        | 825 Middlefield Rd, Redwood    |           |           |             |
|        | City, California 94063, United |           |           |             |
|        | States                         |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      1 | Redwood City, California,      | 49.00%    | 37.485199 | -122.236397 |
|        | United States                  |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      2 | Redwood City Station, 805      | 39.00%    | 37.485439 | -122.231796 |
|        | Veterans Blvd, Redwood City,   |           |           |             |
|        | California 94063, United       |           |           |             |
|        | States                         |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      3 | Cinemark Ave, Markham, Ontario | 39.00%    | 43.887989 |  -79.225441 |
|        | L6B 1E3, Canada                |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      4 | Cinemark Ct, Mulberry, Florida | 39.00%    | 27.934687 |  -81.996933 |
|        | 33860, United States           |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
Please enter your choice by numeric key or (n) to search again: 0
End Point: Palo Alto  
+--------+--------------------------------+-----------+-----------+-------------+
| CHOICE |              NAME              | RELEVANCE | LATITUDE  |  LONGITUDE  |
+--------+--------------------------------+-----------+-----------+-------------+
|      0 | Palo Alto, California, United  | 99.00%    | 37.442200 | -122.163399 |
|        | States                         |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      1 | Palo Alto Battlefield National | 99.00%    | 26.021400 |  -97.480598 |
|        | Historical Park, 7200 PAREDES  |           |           |             |
|        | LINE Rd, Los Fresnos, Texas    |           |           |             |
|        | 78566, United States           |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      2 | Palo Alto Baylands Nature      | 99.00%    | 37.459599 | -122.106003 |
|        | Preserve, 2500 Embarcadero     |           |           |             |
|        | Way, East Palo Alto,           |           |           |             |
|        | California 94303, United       |           |           |             |
|        | States                         |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      3 | Palo Alto University, 1791     | 99.00%    | 37.382301 | -122.188004 |
|        | Arastradero Rd, Palo Alto,     |           |           |             |
|        | California 94304, United       |           |           |             |
|        | States                         |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
|      4 | Palo Alto High School, 50      | 99.00%    | 37.437000 | -122.156998 |
|        | Embarcadero Rd, Palo Alto,     |           |           |             |
|        | California 94306, United       |           |           |             |
|        | States                         |           |           |             |
+--------+--------------------------------+-----------+-----------+-------------+
Please enter your choice by numeric key or (n) to search again: 0
Seat count: 1 or 2 (default 2) 1
+--------+--------+----------+----------+----------------------+--------------------+
| CHOICE |  NAME  | ESTIMATE | CURRENCY | PICKUP ETA (MINUTES) | DURATION (MINUTES) |
+--------+--------+----------+----------+----------------------+--------------------+
|      0 | SELECT | $31-39   | USD      |                  3.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
|      1 | ASSIST | $15-19   | USD      |                 10.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
|      2 | uberXL | $19-24   | USD      |                 12.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
|      3 | BLACK  | $40-50   | USD      |                  5.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
|      4 | SUV    | $53-65   | USD      |                  5.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
|      5 | WAV    | $13-16   | USD      |                  0.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
|      6 | POOL   | $6-8     | USD      |                  9.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
|      7 | uberX  | $15-19   | USD      |                  8.0 |               22.0 |
+--------+--------+----------+----------+----------------------+--------------------+
Please enter the choice of your item or n to cancel
```

### payments
payments allows you to list your payments
```shell
$ uber payments
+------------+--------------------------------------+-------------+----------+
|   METHOD   |                  ID                  | DESCRIPTION | LASTUSED |
+------------+--------------------------------------+-------------+----------+
| visa       | 9a152688-e81c-4a17-91f4-27bde532b7f1 | ***48       | ✔️        |
+------------+--------------------------------------+-------------+----------+
| visa       | 83634d20-9036-4797-87e2-fb8dcf574b7b | ***39       |          |
+------------+--------------------------------------+-------------+----------+
| unknown    | 3c4b8f3c-6924-426f-b837-c3aba3a2eecb |             |          |
+------------+--------------------------------------+-------------+----------+
| mastercard | 90b24751-8414-4e72-be33-d245f42f4be1 | ***31       |          |
+------------+--------------------------------------+-------------+----------+
```
