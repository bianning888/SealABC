package p2pv2

import (
	"github.com/SealSC/SealABC/crypto/signers/signerCommon"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func getPrivateKey(signer signerCommon.ISigner) (pk crypto.PrivKey, err error) {
	sdk := signer.ToStandardKey()
	pk, _, err = crypto.KeyPairFromStdKey(sdk)

	return
}
