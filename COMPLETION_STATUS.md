# completion status
Provides an almost one-to-one mapping of Uber REST API methods to those in this API client.

## Table of contents
- [Rides API](#rides-api)
# Rides API
Uber API Method | API Method | Completion Status | Notes | Description
---|---|---|---|---
GET /authorize|oauth2.Authorize|✔️|Authentication|Allows you to redirect a user to the authorization URL for your application. See samples https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/cmd/uber/main.go#L111-L130 and https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/oauth2/oauth2.go#L209-L257 
POST /token|oauth2.Authorize|✔️|Authentication|The Login endpoint that allows you to authorize your application and get an access token using the authorization code or client credentials grant. See for example, see samples https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/cmd/uber/main.go#L111-L130 and https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/oauth2/oauth2.go#L209-L257 
PATCH /me|client.ApplyPromoCode|✔️||Allows you to apply a promocode to your account. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L202-L214
GET /history|client.ListHistory|✔️||Retrieve the history of your trips. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L45-L80
GET /payment-methods|client.ListPaymentMethods|✔️||Retrieves your payment methods. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L26-L43
GET /places/{place_id}|client.Place|✔️||Retrieves either your HOME or WORK addresses, if set. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L230-L242 and https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L244-L256
PUT /places/{place_id}|client.UpdatePlace|✔️||Updates either your HOME or WORK addresses. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L258-L273 and https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L275-L290
GET /products|client.ListProducts|✔️||Allows you to get a list of products/car options at a location. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L439-L456
GET /products/{product_id}|client.ProductByID|||Retrieves a product/car option by its ID. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L458-L470
GET /estimates/price|client.EstimatePrice|||Returns an estimated price range for each product offered at a given location. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L114-L148
GET /estimates/time|client.EstimateTime|✔️||Returns ETAs for all products currently available at a given location. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L150-L186
GET /requests/estimate|client.UpfrontFare|✔️|Privileged scope, so needs an OAuth2.0 authorized client. This method is needed before you request a ride|Allows retrieve the upfront fare for all products currently available at a given location. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L317-L343
POST /requests|client.RequestRide|✔️|Privileged scope, OAuth2.0 bearer token with the request scope. Requires you to pass in the FareID retrieved from client.UpfrontFare|See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L345-L368
GET /requests/current||✖️|Unimplemented|Retrieve details of an ongoing trip
PATCH /requests/current||✖️|Unimplemented|Update an ongoing trip's destination
DELETE /requests/current||✖️|Unimplemented|Cancel the ongoing trip
GET /requests/{request_id}|||Unimplemented|Retrieve the details of an ongoing or completed trip that was created by your app
PATCH /requests/{request_id}||✖️|Unimplemented|Update the ongoing request's destination using the Ride Request endpoint
DELETE /requests/{request_id}|||Unimplemented|Cancel the ongoing request on behalf of a rider
GET /requests/{request_id}/map|client.OpenMap|✔️||This method is only available after a trip has been accepted by a driver and is in the accepted state|Opens up the map for an trip, to give a visual representation of a request. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L306-L315
GET /requests/{request_id}/receipt|client.RequestReceipt|✔️|A privileged scope, whose output is only available after the requests.receipt_ready webhook notification is sent|The trip receipt may be adjusted after the requests.receipt_ready webhook is sent as finalized receipts can be delayed. See https://github.com/orijtech/uber/blob/1c064b69c7686b21ee5768468f39b900a2c1e8cb/example_test.go#L216-L228

