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
	"errors"
	"log"
)

type ActionableError struct {
	msg       string
	code      int
	action    string
	retryable bool
	signature string
}

var _ error = (*ActionableError)(nil)

func (ae *ActionableError) Error() string {
	if ae == nil {
		return ""
	}

	return ae.msg
}

func (ae *ActionableError) HasAction() bool {
	return ae != nil && len(ae.action) > 0
}

func (ae *ActionableError) Action() string {
	if ae == nil {
		return ""
	}
	return ae.action
}

var (
	ErrProccessingRequest = errors.New("error_")
)

// UberErrors
// * 400 :: Unconfirmed email :: The user hasn't confirmed their email address.
//				 Instruct them to confirm their email by visiting
//				 https://riders.uber.com or within the native mobile
//				 application.
var ErrUnconfirmedEmail = &ActionableError{
	msg:       "unconfirmed email address",
	code:      400,
	signature: "unconfirmed_email",
	action:    "https://riders.uber.com",
}

// * 400 :: error_processing_request :: Error processing the request.
var ErrProcessingRequest = &ActionableError{
	msg:       "encountered a error processing the request",
	code:      400,
	signature: "error_processing_request",
}

// * 400 :: promotions_revoked :: Promotions revoked.
var ErrPromotionsRevoked = &ActionableError{
	msg:       "encountered a error processing the request",
	code:      400,
	signature: "promotions_revoked",
}

// * 400 :: invalid_payment :: The rider's payment method is invalid. The user
//			       must update the billing info. This could include e.g
//			       Android Pay.
var ErrInvalidPayment = &ActionableError{
	msg:       "rider's payment is invalid. Please update billing information",
	code:      400,
	signature: "invalid_payment",
}

// * 400 :: invalid_payment_method :: The provided payment method is not valid.
var ErrInvalidPaymentMethod = &ActionableError{
	msg:       "the provided payment method is not valid",
	code:      400,
	signature: "invalid_payment_method",
}

// * 400 :: outstanding_balance_update_billing :: The user has outstanding balances.
//						  The user must update the billing info.
var ErrOutstandingBalance = &ActionableError{
	msg:       "your account has outstanding balances. Please update billing information",
	code:      400,
	signature: "outstanding_balance_update_billing",
}

// * 400 :: insufficient_balance :: There is insufficient balance on the credit card associated
//				    with the user. The user must update the billing info.
var ErrInsufficientBalance = &ActionableError{
	msg:       "insufficient balance on the credit card associated with your account. Please update billing information",
	code:      400,
	signature: "insufficient_balance",
}

// * 400 :: payment_method_not_allowed :: The payment method is not allowed.
var ErrPaymentMethodNotAllowed = &ActionableError{
	msg:       "the payment method is not allowed",
	code:      400,
	signature: "payment_method_not_allowed",
}

// * 400 :: card_assoc_outstanding_balance :: The user's associated card has an outstanding balance.
//					      The user must update the billing info.
var ErrCardHasOutstandingBalance = &ActionableError{
	msg:       "the associated card has an outstanding balance. Please update billing information",
	code:      400,
	signature: "card_assoc_outstanding_balance",
}

// * 400 :: invalid_mobile_phone_number :: The user's mobile phone number is not supported. We don't
//					   allow phone numbers for some providers that allow the creation
//					   of temporary phone numbers.
var ErrInvalidMobilePhoneNumber = &ActionableError{
	msg:       "the mobile phone number is not supported. We don't allow phone numbers for some providers that allow the creation of temporary phone numbers",
	code:      400,
	signature: "invalid_mobile_phone_number",
}

// * 403 :: forbidden :: This user is forbidden from making a request at this time and should consult
//			 our support team by visiting https://help.uber.com or by emailing support@uber.com
var ErrForbiddenRequest = &ActionableError{
	msg:       "you are forbidden from making a request at this time. Please consult our support team",
	code:      403,
	signature: "forbidden",
	action:    "https://help.uber.com,support@uber.com",
}

// * 403 :: unverified :: The user hasn't confirmed their phone number. Instruct the user to
//			  confirm their mobile phone number within the native mobile app
//			  or by visiting https://riders.uber.com
var ErrUnverified = &ActionableError{
	msg:       "your phone number hasn't yet been confirmed",
	code:      403,
	signature: "unverified",
	action:    "https://riders.uber.com",
}

// * 403 :: verification_required :: The user currently cannot make ride requests through
//				     the API and is advised to use the Uber iOS or Android
//				     rider app to get a ride.
var ErrVerificationRequired = &ActionableError{
	msg:       "you aren't allowed to make ride requests through the API. Please use the Uber iOS or Android rider app to get a ride",
	code:      403,
	signature: "verification_required",
}

// * 403 :: product_not_allowed :: The product being requested is not available to the user.
//				    Have them select a different product to successfully
//				    make a request.
var ErrProductNotAllowed = &ActionableError{
	msg:       "the requested product is not available to you unfortunately. Please select another product",
	code:      403,
	signature: "product_not_allowed",
}

// * 403 :: pay_balance :: The rider has an outstanding balance and must update their
//			   account settings by using the native mobile application or
//			   by visiting https://riders.uber.com
var ErrPayBalance = &ActionableError{
	msg:       "you have an outstanding balance. Please update your account settings",
	code:      403,
	signature: "pay_balance",
	action:    "https://riders.uber.com",
}

// * 403 :: user_not_allowed :: The user is banned and not permitted to request a ride.
var ErrUserNotAllowed = &ActionableError{
	msg:       "unfortunately you are banned and not permitted to request a ride",
	code:      403,
	signature: "user_not_allowed",
}

// * 403 :: too_many_cancellations :: The rider is temporarily blocked due to canceling
//				      too many times.
var ErrTooManyCancellations = &ActionableError{
	msg:       "you are temporarily blocked for canceling too many times",
	code:      403,
	signature: "too_many_cancellations",
}

// * 403 :: missing_national_id :: Certain jurisdictions require Uber users to register
//				   their national ID number or passport number before
//				   taking a ride. If a user receives this error when
//				   booking a trip through the Developer API, they must
//				   enter their national ID number or passport number
//				   through the Uber iOS or Android app.
var ErrMissingNationalID = &ActionableError{
	msg:       "certain jurisdictions require registration of your national ID or passport number before taking a ride. Please enter your national ID or passport number through the Uber iOS or Android app",
	code:      403,
	signature: "missing_national_id",
}

// * 404 :: no_product_found :: An invalid product ID was requested. Retry the API call
//				with a valid product ID.
var ErrNoProductFound = &ActionableError{
	msg:       "an invalid product ID was requested. Retry the API call with a valid product ID",
	code:      404,
	signature: "no_product_found",
}

// * 409 :: missing_payment_method :: The rider must have at least one payment method
//				      on file to request a car. The rider must add a
//				      payment method by using the native mobile application
//				      or by visiting https://riders.uber.com
var ErrMissingPaymentMethod = &ActionableError{
	msg:       "please add at least one payment method on file before requesting a car",
	code:      409,
	signature: "missing_payment_method",
	action:    "https://riders.uber.com",
}

// * 409 :: surge :: Surge pricing is currently in effect for this product. Please have
//		     the user confirm surge pricing by sending them to the surge_confirmation
//		      href described.
var ErrSurge = &ActionableError{
	msg:       "surge pricing is currently in effect for this product. Please confirm the surge pricing first",
	code:      409,
	signature: "surge",
}

// * 409 :: fare_expired :: The fare has expired for the requested product. Please get estimates
//			    again, confirm the new fare, and then re-request.
var ErrFareExpired = &ActionableError{
	msg:       "the fare has expired for the requested product. Please get estimates again, confirm the new fare and then re-request",
	code:      409,
	signature: "fare_expired",
}

// * 409 :: retry_request :: An error has occured when attempting to request a product.
//			     Please reattempt the request on behalf of the user.
var ErrRetryRequest = &ActionableError{
	msg:       "an error has occured when attempting to reques a product. Please retry the request",
	code:      409,
	signature: "retry_request",
	retryable: true,
}

// * 409 :: current_trip_exists :: The user is currently on a trip.
var ErrUserCurrentlyOnTrip = &ActionableError{
	msg:       "the user is currently on a trip",
	code:      409,
	signature: "current_trip_exists",
}

// * 422 :: invalid_fare_id :: This fare id is invalid or expired.
var ErrInvalidFareID = &ActionableError{
	msg:       "the fare id is invalid or expired",
	code:      422,
	signature: "invalid_fare_id",
}

// * 422 :: destination_required :: This product requires setting a destination
//				    for ride requests.
var ErrDestinationRequired = &ActionableError{
	msg:       "this product requires setting a destination",
	code:      422,
	signature: "destination_required",
}

// * 422 :: distance_exceeded :: The distance between start and end locations exceeds 100 miles.
var ErrDistanceExceeded = &ActionableError{
	msg:       "the distance between start and end location exceeds 100 miles",
	code:      422,
	signature: "distance_exceeded",
}

// * 422 :: same_pickup_dropoff :: Pickup and Dropoff can't be the same.
var ErrSamePickupAsDroOff = &ActionableError{
	msg:       "pickup and dropoff cannot be the same",
	code:      422,
	signature: "same_pickup_dropoff",
}

// * 422 :: validation_failed :: The destination is not supported for uberPOOL.
var ErrInvalidUberPoolDestination = &ActionableError{
	msg:       "this destination is not supported for uberPOOL",
	code:      422,
	signature: "validation_failed",
}

// * 422 :: invalid_seat_count :: Number of seats exceeds max capacity.
var ErrInvalidSeatCount = &ActionableError{
	msg:       "number of seats exceeds max capacity",
	code:      422,
	signature: "invalid_seat_count",
}

// * 422 :: outside_service_area :: The destination is not supported by the requested product.
var ErrDestinationOutsideServiceArea = &ActionableError{
	msg:       "the destination is not supported by the requested product",
	code:      422,
	signature: "outside_service_area",
}

// * 500 :: internal_server_error :: An unknown error has occured.
var ErrInternalServerError = &ActionableError{
	msg:       "an unknown error has occured",
	code:      500,
	signature: "internal_server_error",
}

var actionableErrorsIndex map[string]*ActionableError

var actionableErrsList = [...]*ActionableError{
	0:  ErrUnconfirmedEmail,
	1:  ErrProcessingRequest,
	2:  ErrPromotionsRevoked,
	3:  ErrInvalidPayment,
	4:  ErrInvalidPaymentMethod,
	5:  ErrOutstandingBalance,
	6:  ErrInsufficientBalance,
	7:  ErrPaymentMethodNotAllowed,
	8:  ErrCardHasOutstandingBalance,
	9:  ErrInvalidMobilePhoneNumber,
	10: ErrForbiddenRequest,
	11: ErrUnverified,
	12: ErrVerificationRequired,
	13: ErrProductNotAllowed,
	14: ErrPayBalance,
	15: ErrUserNotAllowed,
	16: ErrTooManyCancellations,
	17: ErrMissingNationalID,
	18: ErrNoProductFound,
	19: ErrMissingPaymentMethod,
	20: ErrSurge,
	21: ErrFareExpired,
	22: ErrRetryRequest,
	23: ErrUserCurrentlyOnTrip,
	24: ErrInvalidFareID,
	25: ErrDestinationRequired,
	26: ErrDistanceExceeded,
	27: ErrSamePickupAsDroOff,
	28: ErrInvalidUberPoolDestination,
	29: ErrInvalidSeatCount,
	30: ErrDestinationOutsideServiceArea,
	31: ErrInternalServerError,
}

func init() {
	actionableErrorsIndex = make(map[string]*ActionableError)
	for i, ae := range actionableErrsList {
		_, previouslyIn := actionableErrorsIndex[ae.signature]
		if previouslyIn {
			log.Fatalf("actionableError: %#d signature (%q) already exists", i, ae.signature)
		}
		actionableErrorsIndex[ae.signature] = ae
	}
}

func lookupErrorBySignature(signature string) *ActionableError {
	return actionableErrorsIndex[signature]
}
