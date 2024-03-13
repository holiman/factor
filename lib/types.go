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
	"math/big"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
)

type CLConfig struct {
	Name    string
	Address string
	Headers map[string]string
}

type ELConfig struct {
	Name      string
	Address   string
	Headers   map[string]string
	JwtSecret string
}

type Config struct {
	ElClients []ELConfig
	ClClient  CLConfig
}

type clWithDrawal struct {
	Index     math.HexOrDecimal64 `json:"index"`           // monotonically increasing identifier issued by consensus layer
	Validator math.HexOrDecimal64 `json:"validator_index"` // index of validator associated with withdrawal
	Address   common.Address      `json:"address"`         // target address for withdrawn ether
	Amount    math.HexOrDecimal64 `json:"amount"`          // value of withdrawal in Gwei
}

//go:generate go run github.com/fjl/gencodec -type beaconBlock -field-override beaconBlockMarshaling -out gen_ed.go

// beaconBlock is _almost_ a beacon.ExecutableDataV1, but they have defined the
// json a bit differently.
type beaconBlock struct {
	ParentHash    common.Hash     `json:"parent_hash"    gencodec:"required"`
	FeeRecipient  common.Address  `json:"fee_recipient"  gencodec:"required"`
	StateRoot     common.Hash     `json:"state_root"     gencodec:"required"`
	ReceiptsRoot  common.Hash     `json:"receipts_root"  gencodec:"required"`
	LogsBloom     []byte          `json:"logs_bloom"     gencodec:"required"`
	Random        common.Hash     `json:"prev_randao"    gencodec:"required"`
	Number        uint64          `json:"block_number"   gencodec:"required"`
	GasLimit      uint64          `json:"gas_limit"      gencodec:"required"`
	GasUsed       uint64          `json:"gas_used"       gencodec:"required"`
	Timestamp     uint64          `json:"timestamp"     gencodec:"required"`
	ExtraData     []byte          `json:"extra_data"     gencodec:"required"`
	BaseFeePerGas *big.Int        `json:"base_fee_per_gas" gencodec:"required"`
	BlockHash     common.Hash     `json:"block_hash"     gencodec:"required"`
	Transactions  [][]byte        `json:"transactions"  gencodec:"required"`
	Withdrawals   []*clWithDrawal `json:"withdrawals" gencodec:"optional"`

	BlobGasUsed   uint64 `json:"blob_gas_used" gencodec:"optional"`
	ExcessBlobGas uint64 `json:"excess_blob_gas" gencodec:"optional"`
}

// JSON type overrides for executableData.
type beaconBlockMarshaling struct {
	Number        math.HexOrDecimal64
	GasLimit      math.HexOrDecimal64
	GasUsed       math.HexOrDecimal64
	Timestamp     math.HexOrDecimal64
	BaseFeePerGas *math.HexOrDecimal256
	ExtraData     hexutil.Bytes
	LogsBloom     hexutil.Bytes
	Transactions  []hexutil.Bytes
	BlobGasUsed   math.HexOrDecimal64
	ExcessBlobGas math.HexOrDecimal64
}

// cat head.resp | jq ".data .message .body .execution_payload"
type bellatrixBlock struct {
	Data struct {
		Message struct {
			ParentRoot common.Hash `json:"parent_root"`
			Body       struct {
				ExecutionPayload beaconBlock `json:"execution_payload"'`
			} `json:"body"`
		} `json:"message"`
	} `json:"data"`
}

func (b beaconBlock) toExecutableDataV1() engine.ExecutableData {
	var withdrawals []*types.Withdrawal

	for _, w := range b.Withdrawals {
		ww := &types.Withdrawal{
			Index:     uint64(w.Index),
			Validator: uint64(w.Validator),
			Address:   w.Address,
			Amount:    uint64(w.Amount),
		}
		withdrawals = append(withdrawals, ww)
	}

	resp := engine.ExecutableData{
		Withdrawals:   withdrawals,
		ParentHash:    b.ParentHash,
		FeeRecipient:  b.FeeRecipient,
		StateRoot:     b.StateRoot,
		ReceiptsRoot:  b.ReceiptsRoot,
		LogsBloom:     b.LogsBloom,
		Random:        b.Random,
		Number:        b.Number,
		GasLimit:      b.GasLimit,
		GasUsed:       b.GasUsed,
		Timestamp:     b.Timestamp,
		ExtraData:     b.ExtraData,
		BaseFeePerGas: b.BaseFeePerGas,
		BlockHash:     b.BlockHash,
		Transactions:  b.Transactions,

		ExcessBlobGas: &b.ExcessBlobGas,
		BlobGasUsed:   &b.BlobGasUsed,
	}
	return resp
}
