package mpcrpc

import (
	"encoding/json"

	"github.com/anyswap/mpc-client/log"
)

func buildAcceptTx(txType, keyID, agreeResult string, msgHash, msgContext []string) (string, error) {
	log.Info("buildAcceptTx", "txType", txType, "keyID", keyID, "agreeResult", agreeResult, "msgHash", msgHash, "msgContext", msgContext)
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
	return BuildMPCRawTx(nonce, payload)
}

// DoAcceptSign accept sign
func DoAcceptSign(keyID, agreeResult string, msgHash, msgContext []string) (string, error) {
	rawTX, err := buildAcceptTx("ACCEPTSIGN", keyID, agreeResult, msgHash, msgContext)
	if err != nil {
		return "", err
	}
	return AcceptSign(rawTX)
}

// DoAcceptReqAddr accept request address
func DoAcceptReqAddr(keyID, agreeResult string) (string, error) {
	rawTX, err := buildAcceptTx("ACCEPTREQADDR", keyID, agreeResult, nil, nil)
	if err != nil {
		return "", err
	}
	return AcceptReqAddr(rawTX)
}
