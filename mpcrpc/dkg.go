package mpcrpc

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/anyswap/mpc-client/log"
)

var (
	errDKGWithoutSigs     = errors.New("dkg without enode sigs")
	errDoDKGFailed        = errors.New("do dkg failed")
	errGetDKGResultFailed = errors.New("get dkg result failed")
)

// DoDKG mpc pubkic key generation
func DoDKG(enodeSigs []string) (keyID string, pubkey string, err error) {
	log.Info("mpc DoDKG begin", "enodeSigs", enodeSigs)
	if len(enodeSigs) == 0 {
		return "", "", errDKGWithoutSigs
	}
	keyID, pubkey, err = doDKGImpl(enodeSigs)
	if err != nil {
		log.Error("mpc DoDKG failed", "err", err)
		return "", "", errDoDKGFailed
	}
	log.Info("mpc DoDKG success", "keyID", keyID, "pubkey", pubkey)
	return keyID, pubkey, nil
}

func doDKGImpl(enodeSigs []string) (keyID string, pubkey string, err error) {
	nonce, err := GetReqAddrNonce(mpcUser.String(), mpcRPCAddress)
	if err != nil {
		return "", "", err
	}
	txdata := ReqAddrData{
		TxType:    "REQDCRMADDR",
		GroupID:   mpcSignGroup,
		ThresHold: mpcThreshold,
		Mode:      mpcMode,
		TimeStamp: NowMilliStr(),
		Sigs:      strings.Join(enodeSigs, "|"),
	}
	payload, _ := json.Marshal(txdata)
	rawTX, err := BuildMPCRawTx(nonce, payload)
	if err != nil {
		return "", "", err
	}

	rpcAddr := mpcRPCAddress
	keyID, err = ReqDcrmAddr(rawTX, rpcAddr)
	if err != nil {
		return "", "", err
	}

	pubkey, err = getDKGResult(keyID, rpcAddr)
	if err != nil {
		return "", "", err
	}
	return keyID, pubkey, nil
}

func getDKGResult(keyID, rpcAddr string) (pubkey string, err error) {
	log.Info("start get dkg status", "keyID", keyID)
	var reqAddrStatus *ReqAddrStatus
	i := 0
	timer := time.NewTimer(mpcSignTimeout)
	defer timer.Stop()
LOOP_GET_DKG_STATUS:
	for {
		i++
		select {
		case <-timer.C:
			if err == nil {
				err = errSignTimerTimeout
			}
			break LOOP_GET_DKG_STATUS
		default:
			reqAddrStatus, err = GetReqAddrStatus(keyID, rpcAddr)
			if err == nil {
				pubkey = reqAddrStatus.PubKey
				break LOOP_GET_DKG_STATUS
			}
			switch {
			case errors.Is(err, ErrGetDKGStatusFailed),
				errors.Is(err, ErrGetDKGStatusTimeout):
				break LOOP_GET_DKG_STATUS
			}
		}
		time.Sleep(1 * time.Second)
	}
	if pubkey == "" || err != nil {
		log.Info("get dkg status failed", "keyID", keyID, "retryCount", i, "err", err)
		return "", errGetDKGResultFailed
	}
	log.Info("get dkg status success", "keyID", keyID, "pubkey", pubkey, "retryCount", i)
	return pubkey, nil
}
