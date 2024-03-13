// Copyright 2022 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
)

// remoteCL represents a remote CL client
type remoteCL struct {
	address       string
	client        *http.Client
	customHeaders map[string]string
}

func newRemoteCL(address, name string, customHeaders map[string]string) (*remoteCL, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	return &remoteCL{
		address:       address,
		client:        client,
		customHeaders: customHeaders,
	}, nil
}

func (r *remoteCL) GetHeadBlock() (resp engine.ExecutableData, beaconRoot common.Hash, err error) {
	return r.GetBlock("head")
}
func (r *remoteCL) GetFinalizedBlock() (resp engine.ExecutableData, beaconRoot common.Hash, err error) {
	return r.GetBlock("finalized")
}

// GetBlock fetches a block from the remote CL node. The specifier can be:
// - "finalized",
// - "head",
// - a number
func (r *remoteCL) GetBlock(specifier string) (resp engine.ExecutableData, beaconRoot common.Hash, err error) {
	var internal bellatrixBlock
	u, err := url.JoinPath(r.address, "eth", "v2", "beacon", "blocks", specifier)
	if err != nil {
		return resp, common.Hash{}, err
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return resp, common.Hash{}, err
	}
	for k, v := range r.customHeaders {
		req.Header.Set(k, v)
	}
	res, err := r.client.Do(req)
	if err != nil {
		return resp, common.Hash{}, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resp, common.Hash{}, err
	}
	//fmt.Printf("%v", string(body))
	err = json.Unmarshal(body, &internal)
	if err != nil {
		return resp, common.Hash{}, fmt.Errorf("response code %v, err: %w", res.StatusCode, err)
	}
	beaconRoot = internal.Data.Message.ParentRoot
	resp = internal.Data.Message.Body.ExecutionPayload.toExecutableDataV1()
	return resp, beaconRoot, nil
}
