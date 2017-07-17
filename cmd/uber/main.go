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

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/orijtech/uber/oauth2"
	"github.com/orijtech/uber/v1"

	"github.com/olekukonko/tablewriter"

	"github.com/odeke-em/cli-spinner"
	"github.com/odeke-em/go-utils/fread"
	"github.com/odeke-em/mapbox"
	"github.com/odeke-em/semalim"
)

const repeatSentinel = "n"

var mapboxClient *mapbox.Client

func init() {
	var err error
	mapboxClient, err = mapbox.NewClient()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.SetFlags(0)

	uberClient, err := uber.NewClientFromOAuth2File(os.ExpandEnv("$HOME/.uber/credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	var init bool
	var order bool
	flag.BoolVar(&init, "init", false, "allow a user to authorize this app to make requests on their behalf")
	flag.BoolVar(&order, "order", false, "order an Uber")
	flag.Parse()

	// Make log not print out time info in its prefix.
	log.SetFlags(0)

	switch {
	case init:
		authorize()
	case order:
		spinr := spinner.New(10)
		var startGeocodeFeature, endGeocodeFeature mapbox.GeocodeFeature
		items := [...]struct {
			ft     *mapbox.GeocodeFeature
			prompt string
		}{
			0: {&startGeocodeFeature, "Start Point: "},
			1: {&endGeocodeFeature, "End Point: "},
		}

		linesChan := fread.Fread(os.Stdin)
		for i, item := range items {
			for {
				geocodeFeature, query, err := doSearch(item.prompt, linesChan, "n", spinr)
				if err == nil {
					*item.ft = *geocodeFeature
					break
				}

				switch err {
				case errRepeat:
					fmt.Printf("\033[32mSearching again *\033[00m\n")
					continue
				case errNoMatchFound:
					fmt.Printf("No matches found found for %q. Try again? (y/N) ", query)
					continueResponse := strings.TrimSpace(<-linesChan)
					if strings.HasPrefix(strings.ToLower(continueResponse), "y") {
						continue
					}
					return
				default:
					// Otherwise an unhandled error
					log.Fatalf("%d: search err: %v; prompt=%q", i, err, item.prompt)
				}
			}
		}

		var seatCount int = 2
		for {
			fmt.Printf("Seat count: 1 or 2 (default 2) ")
			seatCountLine := strings.TrimSpace(<-linesChan)
			if seatCountLine == "" {
				seatCount = 2
				break
			} else {
				parsed, err := strconv.ParseInt(seatCountLine, 10, 32)
				if err != nil {
					log.Fatalf("seatCount parsing err: %v", err)
				}
				if parsed >= 1 && parsed <= 2 {
					seatCount = int(parsed)
					break
				} else {
					fmt.Printf("\033[31mPlease enter either 1 or 2!\033[00m\n")
				}
			}
		}

		startCoord := centerToCoord(startGeocodeFeature.Center)
		endCoord := centerToCoord(endGeocodeFeature.Center)
		esReq := &uber.EstimateRequest{
			StartLatitude:  startCoord.Lat,
			StartLongitude: startCoord.Lng,
			EndLatitude:    endCoord.Lat,
			EndLongitude:   endCoord.Lng,
			SeatCount:      seatCount,
		}

		estimates, err := doUberEstimates(uberClient, esReq, spinr)
		if err != nil {
			log.Fatalf("estimate err: %v\n", err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		table.SetHeader([]string{
			"Choice", "Name", "Estimate", "Currency",
			"Pickup time (minutes)", "TripDuration (minutes)",
		})
		for i, est := range estimates {
			et := est.Estimate
			ufare := est.UpfrontFare
			table.Append([]string{
				fmt.Sprintf("%d", i),
				fmt.Sprintf("%s", et.LocalizedName),
				fmt.Sprintf("%s", et.Estimate),
				fmt.Sprintf("%s", et.CurrencyCode),
				fmt.Sprintf("%.1f", ufare.PickupEstimateMinutes),
				fmt.Sprintf("%.1f", et.DurationSeconds/60.0),
			})
		}
		table.Render()

		var estimateChoice *estimateAndUpfrontFarePair

		for {
			fmt.Printf("Please enter the choice of your item or n to cancel ")

			lineIn := strings.TrimSpace(<-linesChan)
			if strings.EqualFold(lineIn, repeatSentinel) {
				return
			}

			choice, err := strconv.ParseUint(lineIn, 10, 32)
			if err != nil {
				log.Fatalf("parsing choice err: %v", err)
			}
			if choice < 0 || choice >= uint64(len(estimates)) {
				log.Fatalf("choice must be >=0 && < %d", len(estimates))
			}
			estimateChoice = estimates[choice]
			break
		}

		if estimateChoice == nil {
			log.Fatal("illogical error, estimateChoice cannot be nil")
		}

		rreq := &uber.RideRequest{
			StartLatitude:  startCoord.Lat,
			StartLongitude: startCoord.Lng,
			EndLatitude:    endCoord.Lat,
			EndLongitude:   endCoord.Lng,
			SeatCount:      seatCount,
			FareID:         string(estimateChoice.UpfrontFare.Fare.ID),
			ProductID:      estimateChoice.Estimate.ProductID,
		}
		spinr.Start()
		rres, err := uberClient.RequestRide(rreq)
		spinr.Stop()
		if err != nil {
			log.Fatalf("requestRide err: %v", err)
		}

		fmt.Printf("\033[33mRide\033[00m\n")
		dtable := tablewriter.NewWriter(os.Stdout)
		dtable.SetHeader([]string{
			"Status", "RequestID", "Driver", "Rating", "Phone", "Shared", "Pickup ETA", "Destination ETA",
		})

		locationDeref := func(loc *uber.Location) *uber.Location {
			if loc == nil {
				loc = new(uber.Location)
			}
			return loc
		}

		dtable.Append([]string{
			fmt.Sprintf("%s", rres.Status),
			rres.RequestID,
			rres.Driver.Name,
			fmt.Sprintf("%d", rres.Driver.Rating),
			fmt.Sprintf("%s", rres.Driver.PhoneNumber),
			fmt.Sprintf("%v", rres.Shared),
			fmt.Sprintf("%.1f", locationDeref(rres.Pickup).ETAMinutes),
			fmt.Sprintf("%.1f", locationDeref(rres.Destination).ETAMinutes),
		})
		dtable.Render()

		vtable := tablewriter.NewWriter(os.Stdout)
		fmt.Printf("\n\033[32mVehicle\033[00m\n")
		vtable.SetHeader([]string{
			"Make", "Model", "License plate", "Picture",
		})
		vtable.Append([]string{
			rres.Vehicle.Make,
			rres.Vehicle.Model,
			rres.Vehicle.LicensePlate,
			rres.Vehicle.PictureURL,
		})
		vtable.Render()
	}
}

func doUberEstimates(uberC *uber.Client, esReq *uber.EstimateRequest, spinr *spinner.Spinner) ([]*estimateAndUpfrontFarePair, error) {
	spinr.Start()
	estimatesPageChan, cancelPaging, err := uberC.EstimatePrice(esReq)
	spinr.Stop()
	if err != nil {
		return nil, err
	}

	var allEstimates []*uber.PriceEstimate
	for page := range estimatesPageChan {
		if page.Err == nil {
			allEstimates = append(allEstimates, page.Estimates...)
		}
		if len(allEstimates) >= 400 {
			cancelPaging()
		}
	}

	spinr.Start()
	defer spinr.Stop()

	jobsBench := make(chan semalim.Job)
	go func() {
		defer close(jobsBench)

		for i, estimate := range allEstimates {
			jobsBench <- &lookupFare{
				client:   uberC,
				id:       i,
				estimate: estimate,
				esReq: &uber.EstimateRequest{
					StartLatitude:  esReq.StartLatitude,
					StartLongitude: esReq.StartLongitude,
					StartPlace:     esReq.StartPlace,
					EndPlace:       esReq.EndPlace,
					EndLatitude:    esReq.EndLatitude,
					EndLongitude:   esReq.EndLongitude,
					SeatCount:      esReq.SeatCount,
					ProductID:      estimate.ProductID,
				},
			}
		}
	}()

	var pairs []*estimateAndUpfrontFarePair
	resChan := semalim.Run(jobsBench, 5)
	for res := range resChan {
		// No ordering required so can just retrieve and add results in
		if retr := res.Value().(*estimateAndUpfrontFarePair); retr != nil {
			pairs = append(pairs, retr)
		}
	}

	return pairs, nil
}

var (
	errNoMatchFound = errors.New("no matches found")
	errRepeat       = errors.New("repeat match")
)

func doSearch(prompt string, linesChan <-chan string, repeatSentinel string, spinr *spinner.Spinner) (*mapbox.GeocodeFeature, string, error) {
	fmt.Printf(prompt)

	query := strings.TrimSpace(<-linesChan)
	if query == "" {
		return nil, query, errRepeat
	}
	spinr.Start()
	matches, err := mapboxClient.LookupPlace(query)
	spinr.Stop()
	if err != nil {
		return nil, query, err
	}

	if len(matches.Features) == 0 {
		return nil, query, errNoMatchFound
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Choice", "Name", "Relevance", "Latitude", "Longitude"})
	table.SetRowLine(true)

	for i, feat := range matches.Features {
		coord := centerToCoord(feat.Center)
		table.Append([]string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%s", feat.PlaceName),
			fmt.Sprintf("%.2f%%", feat.Relevance*100),
			fmt.Sprintf("%f", coord.Lat),
			fmt.Sprintf("%f", coord.Lng),
		})
	}
	table.Render()

	fmt.Printf("Please enter your choice by numeric key or (%v) to search again: ", repeatSentinel)
	lineIn := strings.TrimSpace(<-linesChan)
	if strings.EqualFold(lineIn, repeatSentinel) {
		return nil, query, errRepeat
	}

	choice, err := strconv.ParseUint(lineIn, 10, 32)
	if err != nil {
		return nil, query, err
	}
	if choice < 0 || choice >= uint64(len(matches.Features)) {
		return nil, query, fmt.Errorf("choice must be >=0 && < %d", len(matches.Features))
	}
	return matches.Features[choice], query, nil
}

func input() string {
	var str string
	fmt.Scanln(os.Stdin, &str)
	return str
}

type coord struct {
	Lat, Lng float64
}

func centerToCoord(center []float32) *coord {
	return &coord{Lat: float64(center[1]), Lng: float64(center[0])}
}

func authorize() {
	uberCredsDirPath, err := ensureUberCredsDirExists()
	if err != nil {
		log.Fatal(err)
	}

	scopes := []string{
		oauth2.ScopeProfile, oauth2.ScopeRequest,
		oauth2.ScopeHistory, oauth2.ScopePlaces,
		oauth2.ScopeRequestReceipt, oauth2.ScopeDelivery,
	}

	token, err := oauth2.AuthorizeByEnvApp(scopes...)
	if err != nil {
		log.Fatal(err)
	}

	blob, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}

	credsPath := filepath.Join(uberCredsDirPath, "credentials.json")
	f, err := os.Create(credsPath)
	if err != nil {
		log.Fatal(err)
	}

	f.Write(blob)
	log.Printf("Successfully saved your OAuth2.0 token to %q", credsPath)
}

func ensureUberCredsDirExists() (string, error) {
	wdir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	curDirPath := filepath.Join(wdir, ".uber")
	if err := os.MkdirAll(curDirPath, 0777); err != nil {
		return "", err
	}
	return curDirPath, nil
}

type lookupFare struct {
	id       int
	estimate *uber.PriceEstimate
	esReq    *uber.EstimateRequest
	client   *uber.Client
}

var _ semalim.Job = (*lookupFare)(nil)

func (lf *lookupFare) Id() interface{} {
	return lf.id
}

func (lf *lookupFare) Do() (interface{}, error) {
	upfrontFare, err := lookupUpfrontFare(lf.client, &uber.EstimateRequest{
		StartLatitude:  lf.esReq.StartLatitude,
		StartLongitude: lf.esReq.StartLongitude,
		StartPlace:     lf.esReq.StartPlace,
		EndPlace:       lf.esReq.EndPlace,
		EndLatitude:    lf.esReq.EndLatitude,
		EndLongitude:   lf.esReq.EndLongitude,
		SeatCount:      lf.esReq.SeatCount,
		ProductID:      lf.estimate.ProductID,
	})

	return &estimateAndUpfrontFarePair{Estimate: lf.estimate, UpfrontFare: upfrontFare}, err
}

type estimateAndUpfrontFarePair struct {
	Estimate    *uber.PriceEstimate `json:"estimate"`
	UpfrontFare *uber.UpfrontFare   `json:"upfront_fare"`
}

func lookupUpfrontFare(c *uber.Client, rr *uber.EstimateRequest) (*uber.UpfrontFare, error) {
	// Otherwise it is time to get the estimate of the fare
	return c.UpfrontFare(rr)
}
