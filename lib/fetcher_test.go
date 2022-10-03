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
	"github.com/ethereum/go-ethereum/common"
	"os"
	"testing"
)

func TestParseHead(t *testing.T) {
	data, err := os.ReadFile("./testdata/head.resp")
	if err != nil {
		t.Fatal(err)
	}
	var block bellatrixBlock
	if err := json.Unmarshal(data, &block); err != nil {
		t.Fatal(err)
	}
	payload := block.Data.Message.Body.ExecutionPayload
	payload2 := payload.toExecutableDataV1()
	if have, want := payload2.BlockHash, common.HexToHash("0x1873367cf106a66be0fc94c2165aeebab012dc7e896911b2c0ccfc4eb947e2be"); have != want {
		t.Fatalf("have %#x, want %#x", have, want)
	}
}
