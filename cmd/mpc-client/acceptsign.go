package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
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
			mpcServerFlag,
			mpcKeystoreFlag,
			mpcPasswordFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func acceptSign(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	signInfos, err := mpcrpc.GetCurNodeSignInfo()
	if err != nil {
		log.Error("getCurNodeSignInfo failed", "err", err)
		return err
	}

	var signInfo *mpcrpc.SignInfoData
	for _, info := range signInfos {
		if info != nil && strings.EqualFold(info.Key, keyIDArg) {
			signInfo = info
			break
		}
	}

	if signInfo == nil {
		return errors.New("sign keyID is not found in accept list")
	}

	msgHashes := signInfo.MsgHash
	msgContexts := signInfo.MsgContext
	if len(msgHashes) != 1 {
		return errors.New("wrong message hash length, must have exact one element")
	}
	if len(msgContexts) != 2 {
		return errors.New("wrong message context length, must have exact two elements")
	}

	msgHash := msgHashes[0]
	msgContext := msgContexts[1]

	fmt.Println("message hash is", msgHash)
	fmt.Println("message context is", msgContext)

	// verify message context
	msgContextType := msgContexts[0]
	switch strings.ToLower(msgContextType) {
	case "ethtx":
	case "plaintext":
		hash := crypto.Keccak256Hash([]byte(msgContext))
		if hash != common.HexToHash(msgHash) {
			return errors.New("message hash is not the keccak256 hash of plaintext msgContext")
		}
	default:
		return fmt.Errorf("unknown message context type '%v'", msgContextType)
	}

	// interaction to ask if agree or disagree
	fmt.Printf("\nDo you agree or disagree? (y/n) ")
	var yesno string
	_, err = fmt.Scanln(&yesno)
	if err != nil {
		return fmt.Errorf("get reply answer failed. %w", err)
	}

	agreeResult := "DISAGREE"
	if strings.EqualFold(yesno, "y") || strings.EqualFold(yesno, "yes") {
		agreeResult = "AGREE"
	}

	result, err := mpcrpc.DoAcceptSign(keyIDArg, agreeResult, msgHashes, msgContexts)
	if err != nil {
		log.Error("mpc accept sign failed", "keyID", keyIDArg, "rpcResult", result, "err", err)
		return err
	}
	log.Info("mpc accept sign finished", "keyID", keyIDArg, "agreeResult", agreeResult, "rpcResult", result)
	return nil
}
