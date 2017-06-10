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

type Status string

const (
	// The request is matching to
	// the most efficient available driver.
	StatusProcessing Status = "processing"

	// The request was unfulfilled because
	// no drivers were available.
	StatusNoDriversAvailable Status = "no_drivers_available"

	// The request has been accepted by a driver and
	// is "en route" to the start location
	// (i.e. start_latitude and start_longitude).
	// This state can occur multiple times in case of
	// a driver re-assignment.
	StatusAccepted Status = "accepted"

	// The driver has arrived or will be shortly.
	StatusArriving Status = "arriving"

	// The request is "en route" from the
	// start location to the end location.
	StatusInProgress Status = "in_progress"

	// The request has been canceled by the driver.
	StatusDriverCanceled Status = "driver_canceled"

	// The request has been canceled by the rider.
	StatusRiderCanceled Status = "rider_canceled"

	// The request has been completed by the driver.
	StatusCompleted Status = "completed"

	// The receipt for the trip is ready.
	StatusReceiptReady Status = "ready"
)
