package mpcrpc

import (
	"encoding/json"

	"github.com/anyswap/mpc-client/log"
)

func acceptSign(txType, keyID, agreeResult string, msgHash, msgContext []string) (string, error) {
	log.Info("acceptSign", "txType", txType, "keyID", keyID, "agreeResult", agreeResult, "msgHash", msgHash, "msgContext", msgContext)
	nonce := uint64(0)
	data := AcceptData{
		TxType:  txType,
		Key:     keyID,
		Accept:  agreeResult,
		MsgHash: msgHash,
		//MsgContext: msgContext, // context is verified on top level
		TimeStamp: NowMilliStr(),
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	rawTX, err := BuildMPCRawTx(nonce, payload)
	if err != nil {
		return "", err
	}
	return AcceptSign(rawTX)
}

// DoAcceptSign accept sign
func DoAcceptSign(keyID, agreeResult string, msgHash, msgContext []string) (string, error) {
	return acceptSign("ACCEPTSIGN", keyID, agreeResult, msgHash, msgContext)
}

// DoAcceptReqAddr accept request address
func DoAcceptReqAddr(keyID, agreeResult string, msgHash, msgContext []string) (string, error) {
	return acceptSign("ACCEPTREQADDR", keyID, agreeResult, msgHash, msgContext)
}
