package mpcrpc

import (
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/anyswap/mpc-client/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	errSignTimerTimeout     = errors.New("sign timer timeout")
	errDoSignFailed         = errors.New("do sign failed")
	errSignWithoutPublickey = errors.New("sign without public key")
	errGetSignResultFailed  = errors.New("get sign result failed")
	errWrongSignatureLength = errors.New("wrong signature length")
)

// DoSignOne mpc sign single msgHash with context msgContext
func DoSignOne(signPubkey, msgHash, msgContext string) (keyID string, rsvs []string, err error) {
	return DoSign(signPubkey, []string{msgHash}, []string{msgContext})
}

// DoSign mpc sign msgHash with context msgContext
func DoSign(signPubkey string, msgHash, msgContext []string) (keyID string, rsvs []string, err error) {
	log.Info("mpc DoSign", "msgHash", msgHash, "msgContext", msgContext)
	if signPubkey == "" {
		return "", nil, errSignWithoutPublickey
	}
	keyID, rsvs, err = doSignImpl(signPubkey, msgHash, msgContext)
	if err != nil {
		log.Error("mpc DoSign failed", "err", err)
		return "", nil, errDoSignFailed
	}
	log.Info("mpc DoSign success")
	return keyID, rsvs, nil
}

func doSignImpl(signPubkey string, msgHash, msgContext []string) (keyID string, rsvs []string, err error) {
	nonce, err := GetSignNonce(mpcUser.String(), mpcRPCAddress)
	if err != nil {
		return "", nil, err
	}
	txdata := SignData{
		TxType:     "SIGN",
		PubKey:     signPubkey,
		MsgHash:    msgHash,
		MsgContext: msgContext,
		Keytype:    mpcSignType,
		GroupID:    mpcSignGroup,
		ThresHold:  mpcThreshold,
		Mode:       mpcMode,
		TimeStamp:  NowMilliStr(),
	}
	payload, _ := json.Marshal(txdata)
	rawTX, err := BuildMPCRawTx(nonce, payload)
	if err != nil {
		return "", nil, err
	}

	rpcAddr := mpcRPCAddress
	keyID, err = Sign(rawTX, rpcAddr)
	if err != nil {
		return "", nil, err
	}

	rsvs, err = getSignResult(keyID, rpcAddr)
	if err != nil {
		return "", nil, err
	}
	for _, rsv := range rsvs {
		signature := common.FromHex(rsv)
		if len(signature) != crypto.SignatureLength {
			return "", nil, errWrongSignatureLength
		}
	}
	return keyID, rsvs, nil
}

// GetSignStatusByKeyID get sign status by keyID
func GetSignStatusByKeyID(keyID string) (rsvs []string, err error) {
	return getSignResult(keyID, mpcRPCAddress)
}

func getSignResult(keyID, rpcAddr string) (rsvs []string, err error) {
	log.Info("start get sign status", "keyID", keyID)
	var signStatus *SignStatus
	i := 0
	signTimer := time.NewTimer(mpcSignTimeout)
	defer signTimer.Stop()
LOOP_GET_SIGN_STATUS:
	for {
		i++
		select {
		case <-signTimer.C:
			if err == nil {
				err = errSignTimerTimeout
			}
			break LOOP_GET_SIGN_STATUS
		default:
			signStatus, err = GetSignStatus(keyID, rpcAddr)
			if err == nil {
				rsvs = signStatus.Rsv
				break LOOP_GET_SIGN_STATUS
			}
			switch {
			case errors.Is(err, ErrGetSignStatusFailed),
				errors.Is(err, ErrGetSignStatusTimeout):
				break LOOP_GET_SIGN_STATUS
			}
		}
		time.Sleep(3 * time.Second)
	}
	if len(rsvs) == 0 || err != nil {
		log.Info("get sign status failed", "keyID", keyID, "retryCount", i, "err", err)
		return nil, errGetSignResultFailed
	}
	log.Info("get sign status success", "keyID", keyID, "retryCount", i)
	return rsvs, nil
}

// BuildMPCRawTx build mpc raw tx
func BuildMPCRawTx(nonce uint64, payload []byte) (string, error) {
	tx := types.NewTransaction(
		nonce,             // nonce
		mpcToAddr,         // to address
		big.NewInt(0),     // value
		100000,            // gasLimit
		big.NewInt(80000), // gasPrice
		payload,           // data
	)
	signature, err := crypto.Sign(mpcSigner.Hash(tx).Bytes(), mpcKeyWrapper.PrivateKey)
	if err != nil {
		return "", err
	}
	sigTx, err := tx.WithSignature(mpcSigner, signature)
	if err != nil {
		return "", err
	}
	txdata, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		return "", err
	}
	rawTX := hexutil.Encode(txdata)
	return rawTX, nil
}
