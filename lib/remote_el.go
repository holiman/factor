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
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
)

const errBackoffCount = 6
const backOffPeriod = 2 * time.Minute
const contextDeadline = 6 * time.Second

// remoteEL represents a remote Execution Layer client.
type remoteEL struct {
	name       string
	cli        *rpc.Client
	pauseUntil time.Time
	errCount   int
}

func newRemoteEL(addr, name string, jwtSecret string, customHeaders map[string]string) (*remoteEL, error) {

	var opts []rpc.ClientOption
	if sec := common.HexToHash(jwtSecret); sec != (common.Hash{}) {
		opts = append(opts, rpc.WithHTTPAuth(node.NewJWTAuth(sec)))
	} else {
		log.Warn("Using empty jwt-secret")
	}
	for k, v := range customHeaders {
		opts = append(opts, rpc.WithHeader(k, v))
	}
	client, err := rpc.DialOptions(context.Background(), addr, opts...)
	if err != nil {
		return nil, err
	}
	return &remoteEL{
		name: name,
		cli:  client,
	}, nil
}

func (r *remoteEL) ForkchoiceUpdatedV1(update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	var raw json.RawMessage
	var resp engine.ForkChoiceResponse
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(contextDeadline))
	err := r.cli.CallContext(ctx, &raw, "engine_forkchoiceUpdatedV1", update, payloadAttributes)
	if err != nil {
		r.errCount++
		return resp, err
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return resp, err
	}
	r.errCount = 0
	return resp, nil
}

func (r *remoteEL) NewPayloadV3(params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash) (engine.PayloadStatusV1, error) {
	var (
		raw  json.RawMessage
		resp engine.PayloadStatusV1
	)
	if time.Now().Before(r.pauseUntil) {
		return resp, errors.New("client paused")
	}
	if r.errCount == errBackoffCount {
		log.Info("Pausing client", "EL", r.name, "duration", backOffPeriod)
		r.pauseUntil = time.Now().Add(backOffPeriod)
		r.errCount = 0
		// back off a bit
	}
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(contextDeadline))
	err := r.cli.CallContext(ctx, &raw, "engine_newPayloadV3", params, versionedHashes, beaconRoot)
	if err != nil {
		r.errCount++
		return resp, err
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return resp, err
	}
	r.errCount = 0
	return resp, nil
}

func (r *remoteEL) ExchangeTransitionConfigurationV1(config engine.TransitionConfigurationV1) (*engine.TransitionConfigurationV1, error) {
	if time.Now().Before(r.pauseUntil) {
		return nil, errors.New("client paused")
	}
	if r.errCount == errBackoffCount {
		log.Info("Pausing client", "EL", r.name, "duration", backOffPeriod)
		r.pauseUntil = time.Now().Add(backOffPeriod)
		r.errCount = 0
		// back off a bit
	}
	var raw json.RawMessage
	err := r.cli.CallContext(context.Background(), &raw, "engine_exchangeTransitionConfigurationV1", config)
	if err != nil {
		return nil, err
	}
	var resp engine.TransitionConfigurationV1
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *remoteEL) GetPayloadV1(payloadID engine.PayloadID) (*engine.ExecutableData, error) {
	return nil, errors.New("GetPayloadV1 not supported")
}

func (r *remoteEL) Name() string {
	return r.name
}
