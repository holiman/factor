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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type blockUpdate struct {
	execData   engine.ExecutableData
	beaconRoot common.Hash
}

// fetcher fetches data from the remote CL, and feeds it to the EL sink.
type fetcher struct {
	cl      *remoteCL
	sink    ElApi
	wg      sync.WaitGroup
	closeCh chan bool
	finalCh chan blockUpdate
	headCh  chan blockUpdate
}

func NewFetcher(config CLConfig, sink ElApi) (*fetcher, error) {
	cl, err := newRemoteCL(config.Address, config.Name, config.Headers)
	if err != nil {
		return nil, err
	}
	return &fetcher{
		cl:      cl,
		sink:    sink,
		closeCh: make(chan bool),
		finalCh: make(chan blockUpdate, 10),
		headCh:  make(chan blockUpdate, 10),
	}, nil
}

func (f *fetcher) Start() {
	f.wg.Add(2)
	go f.fetchLoop()
	go f.deliverLoop()
}

func (f *fetcher) Stop() {
	close(f.closeCh)
	f.wg.Wait()
}

// loop runs the fetcher loop, which fetches new heads and finalized blocks
// from the CL node, and emits them over the finalCh and headCh.
func (f *fetcher) fetchLoop() {
	defer f.wg.Done()

	var (
		timer = time.NewTimer(10 * time.Second)
		final engine.ExecutableData
		head  engine.ExecutableData
	)
	defer timer.Stop()

	for {
		if newFinal, beaconRoot, err := f.cl.GetFinalizedBlock(); err != nil {
			log.Error("Failed fetching finalized", "err", err)
			timer.Reset(30 * time.Second)
			select {
			case <-timer.C:
			case <-f.closeCh:
				return
			}
		} else if newFinal.Number != 0 && newFinal.Number != final.Number {
			final = newFinal // New finalized
			log.Info("New final block", "number", final.Number, "hash", final.BlockHash, "beaconRoot", beaconRoot)
			select {
			case f.finalCh <- blockUpdate{final, beaconRoot}:
			default:
			}
		}
		if newHead, beaconRoot, err := f.cl.GetHeadBlock(); err != nil {
			log.Error("Failed fetching head", "err", err)
			timer.Reset(30 * time.Second)
			select {
			case <-timer.C:
			case <-f.closeCh:
				return
			}
		} else if newHead.Number != 0 && newHead.Number != head.Number {
			head = newHead // New head
			log.Info("New head block", "number", head.Number, "hash", head.BlockHash, "beaconRoot", beaconRoot)
			select {
			case f.headCh <- blockUpdate{head, beaconRoot}:
			default:
			}
		}
		timer.Reset(10 * time.Second)
		select {
		case <-timer.C:
		case <-f.closeCh:
			return
		}
	}
}

func (f *fetcher) deliverLoop() {
	defer f.wg.Done()

	var (
		lastHead      common.Hash
		lastFinalized common.Hash
	)
	for {
		select {
		case headUpdate := <-f.headCh:
			head := headUpdate.execData
			lastHead = head.BlockHash
			f.sink.NewPayloadV3(head, []common.Hash{}, &headUpdate.beaconRoot)

			msg := engine.ForkchoiceStateV1{HeadBlockHash: lastHead}
			if lastFinalized != (common.Hash{}) {
				msg.FinalizedBlockHash = lastFinalized
			}
			f.sink.ForkchoiceUpdatedV1(msg, nil)

		case finalizedUpdate := <-f.finalCh:
			finalized := finalizedUpdate.execData
			lastFinalized = finalized.BlockHash

			// Initialize the head block hash using the finalized hash
			// in case no event is received from the head channel.
			if lastHead == (common.Hash{}) {
				lastHead = finalized.BlockHash
			}
			f.sink.ForkchoiceUpdatedV1(engine.ForkchoiceStateV1{
				FinalizedBlockHash: finalized.BlockHash,
				HeadBlockHash:      lastHead,
			}, nil)

		case <-f.closeCh:
			return
		}
	}
}
