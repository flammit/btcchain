// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcchain_test

import (
	"github.com/conformal/btcchain"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"testing"
)

// TestHaveBlock tests the HaveBlock API to ensure proper functionality.
func TestHaveBlock(t *testing.T) {
	// Load up blocks such that there is a side chain.
	// (genesis block) -> 1 -> 2 -> 3 -> 4
	//                          \-> 3a
	testFiles := []string{
		"blk_0_to_4.dat.bz2",
		"blk_3A.dat.bz2",
	}

	var blocks []*btcutil.Block
	for _, file := range testFiles {
		blockTmp, err := loadBlocks(file)
		if err != nil {
			t.Errorf("Error loading file: %v\n", err)
			return
		}
		for _, block := range blockTmp {
			blocks = append(blocks, block)
		}
	}

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("haveblock")
	if err != nil {
		t.Errorf("Failed to setup chain instance: %v", err)
		return
	}
	defer teardownFunc()

	// Since we're not dealing with the real block chain, disable
	// checkpoints and set the coinbase maturity to 1.
	chain.DisableCheckpoints(true)
	btcchain.TstSetCoinbaseMaturity(1)

	for i := 1; i < len(blocks); i++ {
		err = chain.ProcessBlock(blocks[i], false)
		if err != nil {
			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
			return
		}
	}

	// Insert an orphan block.
	if err := chain.ProcessBlock(btcutil.NewBlock(&Block100000), false); err != nil {
		t.Errorf("Unable to process block: %v", err)
		return
	}

	tests := []struct {
		hash string
		want bool
	}{
		// Genesis block should be present (in the main chain).
		{hash: btcwire.GenesisHash.String(), want: true},

		// Block 3a should be present (on a side chain).
		{hash: "00000000474284d20067a4d33f6a02284e6ef70764a3a26d6a5b9df52ef663dd", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506", want: true},

		// Random hashes should not be availble.
		{hash: "123", want: false},
	}

	for i, test := range tests {
		hash, err := btcwire.NewShaHashFromStr(test.hash)
		if err != nil {
			t.Errorf("NewShaHashFromStr: %v", err)
			continue
		}

		result := chain.HaveBlock(hash)
		if result != test.want {
			t.Errorf("HaveBlock #%d got %v want %v", i, result,
				test.want)
			continue
		}
	}
}
