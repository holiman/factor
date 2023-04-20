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

import "github.com/ethereum/go-ethereum/beacon/engine"

type ElApi interface {
	// ForkchoiceUpdatedV1 informs the EL about the most recent head.
	ForkchoiceUpdatedV1(update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error)
	// ExchangeTransitionConfigurationV1 checks the given configuration against the configuration of the node.
	ExchangeTransitionConfigurationV1(config engine.TransitionConfigurationV1) (*engine.TransitionConfigurationV1, error)
	// GetPayloadV1 returns a cached payload by id.
	GetPayloadV1(payloadID engine.PayloadID) (*engine.ExecutableData, error)
	// NewPayloadV1 creates an Eth1 block, inserts it in the chain, and returns the status of the chain.
	NewPayloadV1(params engine.ExecutableData) (engine.PayloadStatusV1, error)
	// Name for the EL, as per configuration.
	Name() string
}
