// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

type UnclaimedDraftAPI struct {
	*hellosign
}

func NewUnclaimedDraftAPI(apiKey string) *UnclaimedDraftAPI {
	return &UnclaimedDraftAPI{newHellosign(apiKey)}
}
