package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var (
	acceptSignCommand = &cli.Command{
		Action:      acceptSign,
		Name:        "acceptsign",
		Usage:       "start accept sign interaction",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			keyIDFlag,
			nonInteractiveFlag,
			agreeSignFlag,
			disagreeSignFlag,
			mpcServerFlag,
			mpcKeystoreFlag,
			mpcPasswordFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func isAllKeyID(keyID string) bool {
	return strings.EqualFold(keyID, "all")
}

func isValidKeyID(keyID string, interactiveMode bool) bool {
	if strings.EqualFold(common.HexToHash(keyID).String(), keyID) {
		return true
	}
	return !interactiveMode && isAllKeyID(keyID)
}

func acceptSign(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	mpcCfg.NeedKeyStore = true
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	keyID := ctx.String(keyIDFlag.Name)
	interactiveMode := !ctx.Bool(nonInteractiveFlag.Name)
	if !isValidKeyID(keyID, interactiveMode) {
		return fmt.Errorf("wrong keyID '%v'", keyID)
	}

	if !interactiveMode {
		isAgree := ctx.Bool(agreeSignFlag.Name) && !ctx.Bool(disagreeSignFlag.Name) // disagree first
		agreeResult := getAgreeResult(isAgree)
		return doAcceptSignNoninteractively(keyID, agreeResult)
	}

	signInfo, err := getSignInfoByKeyID(keyID)
	if err != nil {
		return err
	}

	if len(signInfo.MsgContext) == 0 {
		return errors.New("empty message context")
	}

	// verify message context
	msgContextType := signInfo.MsgContext[0]
	switch strings.ToLower(msgContextType) {
	case "ethtx":
		err = verifyEthTxSignInfo(signInfo)
	case "plaintext":
		err = verifyPlainTextSignInfo(signInfo)
	default:
		return fmt.Errorf("unknown message context type '%v'", msgContextType)
	}
	if err != nil {
		log.Error("message context is unresolvable", "err", err)
		log.Info("please check the above message context manually.")
		isContinue := askForReply("Do you still want to continue?")
		if !isContinue {
			return err
		}
	}

	isAgree := askForReply("Do you agree this sign?")
	agreeResult := getAgreeResult(isAgree)
	return doAcceptSign(keyID, agreeResult, signInfo.MsgHash, signInfo.MsgContext)
}

func askForReply(prompt string) bool {
	fmt.Printf("\n%s (y/n) ", prompt)
	var reply string
	_, err := fmt.Scanln(&reply)
	if err != nil {
		log.Fatal("get reply failed", "err", err)
	}
	return strings.EqualFold(reply, "y") || strings.EqualFold(reply, "yes")
}

func verifyEthTxSignInfo(signInfo *mpcrpc.SignInfoData) (err error) {
	msgHashes := signInfo.MsgHash
	msgContexts := signInfo.MsgContext
	if len(msgHashes) != 1 {
		return errors.New("wrong message hash length, must have exact one element")
	}
	if len(msgContexts) < 3 {
		return errors.New("wrong message context length, must have at least three elements")
	}
	msgHash := msgHashes[0]
	msgContext := msgContexts[1]
	chainIDStr := msgContexts[2]
	chainID, ok := new(big.Int).SetString(chainIDStr, 0)
	if !ok {
		return fmt.Errorf("wrong block chainID '%v'", chainIDStr)
	}
	var rawTx types.Transaction
	err = json.Unmarshal([]byte(msgContext), &rawTx)
	if err != nil {
		return fmt.Errorf("json unmarshal msgContext to ethtx failed. %w", err)
	}
	chainSigner := types.NewEIP155Signer(chainID)
	calcedHash := chainSigner.Hash(&rawTx)
	return checkMessageHash(calcedHash, msgHash)
}

func verifyPlainTextSignInfo(signInfo *mpcrpc.SignInfoData) (err error) {
	msgHashes := signInfo.MsgHash
	msgContexts := signInfo.MsgContext
	if len(msgHashes) != 1 {
		return errors.New("wrong message hash length, must have exact one element")
	}
	if len(msgContexts) < 2 {
		return errors.New("wrong message context length, must have at least two elements")
	}
	msgHash := msgHashes[0]
	msgContext := msgContexts[1]
	calcedHash := crypto.Keccak256Hash([]byte(msgContext))
	return checkMessageHash(calcedHash, msgHash)
}

func checkMessageHash(calcedHash common.Hash, msgHash string) error {
	if calcedHash == common.HexToHash(msgHash) {
		return nil
	}
	return fmt.Errorf("check message hash failed. msgHash=%v, calcHash=%v", msgHash, calcedHash.String())
}

func getSignInfoByKeyID(keyID string) (signInfo *mpcrpc.SignInfoData, err error) {
	signInfos, err := mpcrpc.GetCurNodeSignInfo()
	if err != nil {
		log.Error("getCurNodeSignInfo failed", "err", err)
		return nil, err
	}

	for _, info := range signInfos {
		if info != nil && strings.EqualFold(info.Key, keyID) {
			signInfo = info
			break
		}
	}

	if signInfo == nil {
		return nil, errors.New("sign keyID is not found in accept list")
	}

	fmt.Println("message hash is", signInfo.MsgHash)
	fmt.Println("message context is", signInfo.MsgContext)

	return signInfo, nil
}

func getAgreeResult(isAgree bool) string {
	if isAgree {
		return "AGREE"
	}
	return "DISAGREE"
}

func doAcceptSignNoninteractively(keyID, agreeResult string) (err error) {
	if !isAllKeyID(keyID) {
		signInfo, errt := getSignInfoByKeyID(keyID)
		if errt != nil {
			return errt
		}
		return doAcceptSign(keyID, agreeResult, signInfo.MsgHash, signInfo.MsgContext)
	}

	signInfos, err := mpcrpc.GetCurNodeSignInfo()
	if err != nil {
		log.Error("getCurNodeSignInfo failed", "err", err)
		return err
	}

	for _, info := range signInfos {
		go func(signInfo *mpcrpc.SignInfoData) {
			errt := doAcceptSign(signInfo.Key, agreeResult, signInfo.MsgHash, signInfo.MsgContext)
			if errt != nil {
				log.Warn("accept sign failed", "signInfo", signInfo, "agreeResult", agreeResult, "err", err)
			}
		}(info)
	}
	return nil
}

func doAcceptSign(keyID, agreeResult string, msgHashes, msgContexts []string) (err error) {
	result, err := mpcrpc.DoAcceptSign(keyID, agreeResult, msgHashes, msgContexts)
	if err != nil {
		log.Error("mpc accept sign failed", "keyID", keyID, "rpcResult", result, "err", err)
		return err
	}
	log.Info("mpc accept sign finished", "keyID", keyID, "agreeResult", agreeResult, "rpcResult", result)
	return nil
}
