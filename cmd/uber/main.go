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
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/orijtech/mapbox"
	"github.com/orijtech/uber/oauth2"
	"github.com/orijtech/uber/v1"

	"github.com/olekukonko/tablewriter"

	"github.com/odeke-em/cli-spinner"
	"github.com/odeke-em/command"
	"github.com/odeke-em/go-utils/fread"
	"github.com/odeke-em/semalim"
)

const repeatSentinel = "n"

var mapboxClient *mapbox.Client

type initCmd struct {
}

var _ command.Cmd = (*initCmd)(nil)

type paymentsCmd struct {
}

var _ command.Cmd = (*paymentsCmd)(nil)

func (p *paymentsCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (p *paymentsCmd) Run(args []string, defaults map[string]*flag.Flag) {
	credsPath := credsMustExist()
	client, err := uberClientFromFile(credsPath)
	exitIfErr(err)

	listings, err := client.ListPaymentMethods()
	exitIfErr(err)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)

	table.SetHeader([]string{
		"Method", "ID", "Description", "LastUsed",
	})

	for _, method := range listings.Methods {
		lastUsedTok := ""
		if method.ID == listings.LastUsedID {
			lastUsedTok = "✔️"
		}
		table.Append([]string{
			fmt.Sprintf("%s", method.PaymentMethod),
			method.ID,
			method.Description,
			lastUsedTok,
		})
	}
	table.Render()
}

func (a *initCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
}

const (
	uberCredsDir       = ".uber"
	credentialsOutFile = "credentials.json"
)

func (a *initCmd) Run(args []string, defaults map[string]*flag.Flag) {
	uberCredsDirPath, err := ensureUberCredsDirExists()
	if err != nil {
		exitIfErr(fmt.Errorf("init: os.MkdirAll(%q) err=%q", err))
	}

	scopes := []string{
		oauth2.ScopeProfile, oauth2.ScopeRequest,
		oauth2.ScopeHistory, oauth2.ScopePlaces,
		oauth2.ScopeRequestReceipt, oauth2.ScopeDelivery,

		// To allow for driver information retrieval
		oauth2.ScopePartnerAccounts,
		oauth2.ScopePartnerPayments,
		oauth2.ScopePartnerTrips,

		oauth2.ScopeAllTrips,
	}

	token, err := oauth2.AuthorizeByEnvApp(scopes...)
	if err != nil {
		log.Fatal(err)
	}

	blob, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}

	credsPath := filepath.Join(uberCredsDirPath, credentialsOutFile)
	f, err := os.Create(credsPath)
	if err != nil {
		log.Fatal(err)
	}

	f.Write(blob)
	log.Printf("Successfully saved your OAuth2.0 token to %q", credsPath)
}

func uberClientFromFile(path string) (*uber.Client, error) {
	return uber.NewClientFromOAuth2File(path)
}

type historyCmd struct {
	maxPage      int
	limitPerPage int
	noPrompt     bool
	pageOffset   int
	throttleStr  string
}

var _ command.Cmd = (*historyCmd)(nil)

func (h *historyCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.IntVar(&h.maxPage, "max-page", 4, "the maximum number of pages to show")
	fs.BoolVar(&h.noPrompt, "no-prompt", false, "if set, do not prompt")
	fs.IntVar(&h.limitPerPage, "limit-per-page", 0, "limits the number of items retrieved per page")
	fs.IntVar(&h.pageOffset, "page-offset", 0, "positions where to start pagination from")
	fs.StringVar(&h.throttleStr, "throttle", "", "the throttle duration e.g 8s, 10m, 7ms as per https://golang.org/pkg/time/#ParseDuration")
	return fs
}

func credsMustExist() string {
	wdir, err := os.Getwd()
	if err != nil {
		exitIfErr(fmt.Errorf("credentials: os.Getwd err=%q", err))
	}

	fullPath := filepath.Join(wdir, uberCredsDir, credentialsOutFile)
	_, err = os.Stat(fullPath)
	exitIfErr(err)

	return fullPath
}

func (h *historyCmd) Run(args []string, defaults map[string]*flag.Flag) {
	credsPath := credsMustExist()
	client, err := uberClientFromFile(credsPath)
	exitIfErr(err)

	// If an invalid duration is passed, it'll return
	// the zero value which is the same as passing in nothing.
	throttle, _ := time.ParseDuration(h.throttleStr)

	pagesChan, _, err := client.ListHistory(&uber.Pager{
		LimitPerPage: int64(h.limitPerPage),
		MaxPages:     int64(h.maxPage),
		StartOffset:  int64(h.pageOffset),

		ThrottleDuration: throttle,
	})

	exitIfErr(err)

	for page := range pagesChan {
		if page.Err != nil {
			fmt.Printf("Page: #%d err: %v\n", page.PageNumber, page.Err)
			continue
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		table.SetHeader([]string{
			"Trip #", "City", "Date",
			"Duration", "Miles", "RequestID",
		})
		fmt.Printf("\n\033[32mPage: #%d\033[00m\n", page.PageNumber+1)
		for i, trip := range page.Trips {
			startCity := trip.StartCity
			startDate := time.Unix(trip.StartTimeUnix, 0)
			endDate := time.Unix(trip.EndTimeUnix, 0)
			table.Append([]string{
				fmt.Sprintf("%d", i+1),
				fmt.Sprintf("%s", startCity.Name),
				fmt.Sprintf("%s", startDate.Format("2006/01/02 15:04:05 MST")),
				fmt.Sprintf("%s", endDate.Sub(startDate)),
				fmt.Sprintf("%.3f", trip.DistanceMiles),
				fmt.Sprintf("%s", trip.RequestID),
			})
		}
		table.Render()
	}
}

type profileCmd struct {
}

var _ command.Cmd = (*profileCmd)(nil)

func (p *profileCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (p *profileCmd) Run(args []string, defaults map[string]*flag.Flag) {
	credsPath := credsMustExist()
	uberClient, err := uberClientFromFile(credsPath)
	exitIfErr(err)

	myProfile, err := uberClient.RetrieveMyProfile()
	if err != nil {
		log.Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader([]string{"Key", "Value"})

	blob, err := json.Marshal(myProfile)
	if err != nil {
		log.Fatalf("serializing profile: %v", err)
	}
	kvMap := make(map[string]interface{})
	if err := json.Unmarshal(blob, &kvMap); err != nil {
		log.Fatalf("deserializing kvMap: %v", err)
	}

	keys := make([]string, 0, len(kvMap))
	for key := range kvMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := kvMap[key]
		table.Append([]string{
			key, fmt.Sprintf("%v", value),
		})
	}
	table.Render()
}

type orderCmd struct {
}

var _ command.Cmd = (*orderCmd)(nil)

func (o *orderCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (o *orderCmd) Run(args []string, defaults map[string]*flag.Flag) {
	credsPath := credsMustExist()
	uberClient, err := uberClientFromFile(credsPath)
	exitIfErr(err)

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
		"Pickup ETA (minutes)", "Duration (minutes)",
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

func main() {
	// Make log not print out time information as a prefix
	log.SetFlags(0)

	var err error
	mapboxClient, err = mapbox.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	command.On("init", "authorizes and initializes your Uber account", &initCmd{}, nil)
	command.On("order", "order your uber", &orderCmd{}, nil)
	command.On("history", "view your trip history", &historyCmd{}, nil)
	command.On("payments", "list your payments methods", &paymentsCmd{}, nil)
	command.On("profile", "details about your profile", &profileCmd{}, nil)

	command.ParseAndRun()
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
	matches, err := mapboxClient.LookupPlace(context.Background(), query)
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

func ensureUberCredsDirExists() (string, error) {
	wdir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	curDirPath := filepath.Join(wdir, uberCredsDir)
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
