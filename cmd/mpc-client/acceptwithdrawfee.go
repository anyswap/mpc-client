package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var (
	acceptWithdrawFeeCommand = &cli.Command{
		Action:      acceptWithdrawFee,
		Name:        "acceptwithdrawfee",
		Usage:       "start accept withdraw fee in batch",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			senderAddrFlag,
			mpcServerFlag,
			mpcKeystoreFlag,
			mpcPasswordFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}

	withdrawFeeSender string
)

func acceptWithdrawFee(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	mpcCfg.NeedKeyStore = true
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}
	withdrawFeeSender = ctx.String(senderAddrFlag.Name)
	if withdrawFeeSender == "" {
		return errors.New("must specify withdraw fee sender (with --from option)")
	}
	log.Infof("withdraw fee sender is %v", withdrawFeeSender)

	signInfos, err := mpcrpc.GetCurNodeSignInfo(0)
	if err != nil {
		log.Error("getCurNodeSignInfo failed", "err", err)
		return err
	}

	var loop uint64
	for {
		loop++
		log.Infof("start accept loop %v", loop)
		for _, info := range signInfos {
			keyID := info.Key

			isAgree, isIgnore, errf := verifyWithdrawFeeSignInfo(info)
			if isIgnore {
				log.Debug("ignore sign info", "keyID", keyID, "err", errf)
				continue
			}
			if !isAgree {
				log.Warn("diagree sign info", "keyID", keyID, "err", errf)
			}

			agreeResult := getAgreeResult(isAgree)
			return doAcceptSign(keyID, agreeResult, info.MsgHash, info.MsgContext)
		}
		time.Sleep(5 * time.Second)
	}
}

func verifyWithdrawFeeSignInfo(info *mpcrpc.SignInfoData) (isAgree, isIgnore bool, err error) {
	msgHashes := info.MsgHash
	msgContexts := info.MsgContext

	isIgnore = true

	if len(msgHashes) != 1 {
		return isAgree, isIgnore, errors.New("mismatch message hash length")
	}

	if len(msgContexts) != 4 {
		return isAgree, isIgnore, errors.New("mismatch message context length")
	}

	msgHash := msgHashes[0]
	msgContextType := msgContexts[0]
	msgContext := msgContexts[1]
	chainIDStr := msgContexts[2]
	signatureHex := msgContexts[3]

	// verify message context
	if strings.ToLower(msgContextType) != "withdrawfee" {
		return isAgree, isIgnore, errors.New("mismatch message context type")
	}

	chainID, ok := new(big.Int).SetString(chainIDStr, 0)
	if !ok {
		return isAgree, isIgnore, fmt.Errorf("wrong chainID '%v'", chainIDStr)
	}

	var rawTx types.Transaction
	err = json.Unmarshal([]byte(msgContext), &rawTx)
	if err != nil {
		return isAgree, isIgnore, fmt.Errorf("json unmarshal msgContext failed. %w", err)
	}

	log.Printf("the sign is sending the following tx to block chain (chainID: %v)", chainIDStr)
	if errf := printTx(&rawTx, true); errf != nil {
		log.Warn("print transaction failed", "err", errf)
	}
	parseEthTx(&rawTx)

	// the sign info message context is right, will not ignore it from now on
	isIgnore = false

	chainSigner := types.NewEIP155Signer(chainID)
	calcedHash := chainSigner.Hash(&rawTx)
	err = checkMessageHash(calcedHash, msgHash)
	if err != nil {
		return isAgree, isIgnore, err
	}

	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return isAgree, isIgnore, fmt.Errorf("wrong signature format: %w", err)
	}

	recoveredPub, err := crypto.Ecrecover(common.HexToHash(msgHash).Bytes(), signature)
	if err != nil {
		return isAgree, isIgnore, fmt.Errorf("recover signature failed: %w", err)
	}
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if !strings.EqualFold(withdrawFeeSender, recoveredAddr.String()) {
		return isAgree, isIgnore, fmt.Errorf("mismatch signature signer: %v", recoveredAddr.String())
	}

	return true, isIgnore, nil
}
