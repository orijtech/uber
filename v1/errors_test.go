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

package uber

import (
	"testing"
)

func TestErrorLookups(t *testing.T) {
	for i, ae := range actionableErrsList {
		signature := ae.signature
		retrAE := lookupErrorBySignature(signature)
		if retrAE != ae {
			t.Errorf("#%d: %q signature lookup returned (%#v) wanted (%#v)", i, signature, retrAE, ae)
		}
	}
}

func TestRandomSignatures(t *testing.T) {
	tests := [...]struct {
		signature string
		want      *ActionableError
	}{
		0: {"unconfirmed_email", ErrUnconfirmedEmail},
		1: {"xyz", nil},
		2: {"too_many_cancellations", ErrTooManyCancellations},
		3: {"missing_national_id", ErrMissingNationalID},
	}

	for i, tt := range tests {
		got := lookupErrorBySignature(tt.signature)
		want := tt.want
		if got != want {
			t.Errorf("#%d: got=%v want=%v", i, got, want)
		}
	}
}
