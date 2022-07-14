package p2pv2

import (
	"encoding/hex"
	"github.com/SealSC/SealABC/crypto/signers/ed25519"
	"testing"
)

func TestPK(t *testing.T) {
	sk, _ := hex.DecodeString("9ebfd0034224d931b27a44b090cc2f1f4ed6b6284ffd4478c660d37ad4d63844e3b1a8afe0c8b6ef66ac5996b73c6352adbd68c9c29be6f464daf28982d3fe4c")
	selfSigner, _ := ed25519.SignerGenerator.FromRawPrivateKey(sk)

	key, err := getPrivateKey(selfSigner)

	if err != nil {
		t.Log(err)
		return
	}

	raw, _ := key.Raw()
	t.Log(hex.EncodeToString(raw))
}
