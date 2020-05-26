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
	"SealABC/storage/db/dbInterface/kvDatabase"
	"math/big"
)

var bigZero = big.NewInt(0)

func (l Ledger) getBalance(addr []byte, hash []byte, cache txResultCache) (*big.Int, error) {
	addrStr := string(addr)
	if cache[addrStr] != nil {
		return cache[addrStr].val, nil
	}

	balance, err := l.balanceOf(addr, hash)
	if err == nil {
		cache[addrStr] = &txResultCacheData{
			val:balance,
		}
	}

	return balance, err
}

func (l Ledger) verifyTransfer(tx Transaction, cache txResultCache) ([]StateData, error) {
	if tx.Type != TxType.Transfer.String() {
		return nil, Errors.InvalidTransactionType
	}

	_, err := tx.DataSeal.Verify(tx.getData(), l.CryptoTools.HashCalculator)
	if err != nil {
		return nil, Errors.DBError.NewErrorWithNewMessage(err.Error())
	}

	assetsHash := l.genesisAssets.getHash()
	fromBalance, err := l.getBalance(tx.From, assetsHash, cache)
	if err != nil {
		return nil, err
	}

	toBalance, err := l.getBalance(tx.To, assetsHash, cache)
	if err != nil {
		return nil, err
	}

	if fromBalance.Cmp(bigZero) <= 0 {
		return nil, Errors.InsufficientBalance
	}

	amount, valid := big.NewInt(0).SetString(string(tx.Value), 10)
	if !valid {
		return nil, Errors.InvalidTransferValue
	}

	if amount.Sign() < 0 {
		return nil, Errors.NegativeTransferValue
	} else if amount.Sign() == 0 {
		return nil, nil
	}

	if fromBalance.Cmp(amount) < 0 {
		return nil, Errors.InsufficientBalance
	}

	fromBalance.Sub(fromBalance, amount)
	toBalance.Add(toBalance, amount)

	statusToChange := []StateData{
		{
			Key: tx.From,
			Val: fromBalance.Bytes(),
		},

		{
			Key: tx.To,
			Val: toBalance.Bytes(),
		},
	}
	return statusToChange, nil
}

func (l Ledger) preTransfer(tx Transaction, cache txResultCache) ([]StateData, txResultCache, error) {
	if cache == nil {
		cache = txResultCache{}
	}

	statusToChange, err := l.verifyTransfer(tx, cache)

	if err != nil {
		return nil, cache, err
	}

	return statusToChange, cache, err
}
//
//func (l Ledger) batchPreTransferActuator(txList []Transaction) error {
//	balanceCache := map[string] *big.Int{}
//
//	for _, tx := range txList {
//		statusToChange, _, err := l.preTransfer(tx, balanceCache)
//		if err != nil {
//			errEl, ok := err.(enum.ErrorElement)
//			if !ok {
//				return err
//			}
//
//			tx.TransactionResult.Success = false
//			tx.TransactionResult.ErrorCode = errEl.Code()
//		} else {
//			tx.TransactionResult.Success = true
//			tx.TransactionResult.NewStatus = statusToChange
//		}
//	}
//
//	return nil
//}

func (l Ledger) batchTransferActuator(txList []Transaction) (err error) {

	var statusKVList []kvDatabase.KVItem

	for _, tx := range txList {
		for _, s := range tx.TransactionResult.NewStatus {
			statusKVList = append(statusKVList, kvDatabase.KVItem {
				Key:    s.Key,
				Data:   s.Val,
				Exists: true,
			})
		}
	}

	err = l.Storage.BatchPut(statusKVList)
	return
}
