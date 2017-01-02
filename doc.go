// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

/*

Package hellosign implements various API clients for the HelloSign platform.

Charges

The creation of live signature requests is not free and requires a paid API plan (https://www.hellosign.com/api/pricing). The
API will return HTTP 402 if such requests are made without a proper plan. The api can still be used for testing purposes
by setting TestMode in parameters to endpoints that create signature requests.

Rate Limits

By default, you can make up to 2000 requests per hour for standard API requests, and 500 requests per hour for higher tier API requests.
In test mode, you can do 50 requests per hour. Exceptions can be made for customers with higher volumes.

*/
package hellosign
