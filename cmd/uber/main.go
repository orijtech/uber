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
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/orijtech/uber/oauth2"
)

func main() {
	var init bool
	flag.BoolVar(&init, "init", false, "allow a user to authorize this app to make requests on their behalf")
	flag.Parse()

	// Make log not print out time info in its prefix.
	log.SetFlags(0)

	switch {
	case init:
		authorize()
	}
}

func authorize() {
	uberCredsDirPath, err := ensureUberCredsDirExists()
	if err != nil {
		log.Fatal(err)
	}

	scopes := []string{
		oauth2.ScopeProfile, oauth2.ScopeRequest,
		oauth2.ScopeHistory, oauth2.ScopePlaces,
		oauth2.ScopeRequestReceipt,
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
