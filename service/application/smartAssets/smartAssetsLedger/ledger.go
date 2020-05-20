/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package smartAssetsLedger

import (
	"SealABC/common/utility/serializer/structSerializer"
	"SealABC/crypto"
	"SealABC/crypto/signers/signerCommon"
	"SealABC/dataStructure/enum"
	"SealABC/metadata/block"
	"SealABC/metadata/blockchainRequest"
	"SealABC/metadata/seal"
	"SealABC/storage/db/dbInterface/kvDatabase"
	"encoding/json"
	"errors"
	"sync"
)

type Ledger struct {
	txPool        [] *blockchainRequest.Entity
	txHashRecord  map[string] bool
	txPoolLimit   int
	clientTxCount map[string] int
	clientTxLimit int

	operateLock sync.RWMutex
	poolLock    sync.Mutex


	genesisAssets BaseAssets
	genesisSigner signerCommon.ISigner

	CryptoTools   crypto.Tools
	Storage       kvDatabase.IDriver
}

func Load() {
	enum.SimpleBuild(&StoragePrefixes)
	enum.SimpleBuild(&TxType)
}

func (l *Ledger) LoadGenesisAssets(creatorKey interface{}, assets BaseAssetsData) error  {

	supply := assets.Supply
	assets.Supply = "0"

	if creatorKey == nil {
		newSigner, err := l.CryptoTools.SignerGenerator.NewSigner(nil)
		if err != nil {
			return err
		}

		creatorKey = newSigner.PrivateKeyBytes()
	}

	metaBytes, _ := structSerializer.ToMFBytes(assets)
	metaSeal := seal.Entity{}
	err := metaSeal.Sign(metaBytes, l.CryptoTools, creatorKey)
	if err != nil {
		return err
	}

	gAssets, exists, _ := l.getAssetsByHash(metaSeal.Hash)

	if !exists {
		assets.Supply = supply
		issuedBytes, _ := structSerializer.ToMFBytes(assets)
		issuedSeal := seal.Entity{}
		err = issuedSeal.Sign(issuedBytes, l.CryptoTools, creatorKey)
		if err != nil {
			return err
		}

		gAssets = &BaseAssets{
			BaseAssetsData: assets,
			IssuedSeal:     issuedSeal,
			MetaSeal:       metaSeal,
		}

		err = l.storeAssets(*gAssets)
		return err
	}

	return nil
}

func (l *Ledger) AddTx(req blockchainRequest.Entity) error {
	tx := Transaction{}
	err := json.Unmarshal(req.Data, &tx)
	if err != nil {
		return err
	}

	if tx.Type != req.RequestAction {
		return errors.New("transaction type is not equal to block request action")
	}
	
	valid, err := tx.verify(l.CryptoTools.HashCalculator)
	if !valid {
		return err
	}

	//todo: check balance to ensure client has enough base-assets to pay transaction basic fee

	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	client := string(tx.Seal.SignerPublicKey)
	clientTxCount := l.clientTxCount[client]
	if clientTxCount >= l.clientTxLimit {
		return errors.New("reach transaction count limit")
	}

	if len(l.txPool) >= l.txPoolLimit {
		return errors.New("reach transaction pool limit")
	}

	_, exists, _ := l.getTxFromStorage(tx.Seal.Hash)
	if exists {
		return errors.New("duplicate history transaction")
	}

	txHash := string(tx.Seal.Hash)
	if l.txHashRecord[txHash] {
		return errors.New("duplicate pending transaction")
	}

	l.txPool = append(l.txPool, &req)
	l.clientTxCount[client] = clientTxCount + 1

	return nil
}

func (l Ledger) PreExecute(txList []Transaction, blockHeader block.Header) (result []byte, err error) {
	//only check signature & duplicate
	txHash := map[string] bool{}
	for _, tx := range txList {
		_, err = tx.verify(l.CryptoTools.HashCalculator)
		if err != nil {
			break
		}

		hash := string(tx.Seal.Hash)
		if _, exists := txHash[hash]; !exists {
			txHash[hash] = true
		} else {
			err = errors.New("duplicate transaction")
		}
	}

	return
}

func (l Ledger) Execute(txList []Transaction, blockHeader block.Header) (result []byte, err error) {
	return
}

func (l Ledger) GetTransactionsFromPool() (txList []blockchainRequest.Entity, count uint32) {
	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	count = uint32(len(l.txPool))
	if count == 0 {
		return
	}

	for _, i := range l.txPool {
		tx := Transaction{}

		//will never unmarshal fail because we marshal correctly when it's appending to the pool
		_ = json.Unmarshal(i.Data, &tx)


	}

	return
}
