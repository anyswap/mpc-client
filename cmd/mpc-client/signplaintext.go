package main

import (
	"errors"
	"fmt"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var (
	signPlainTextCommand = &cli.Command{
		Action:      signPlainText,
		Name:        "signplaintext",
		Usage:       "sign plain text",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			pubkeyFlag,
			msgHashFlag,
			msgContextFlag,
			signMessageFlag,
			gidFlag,
			thresholdFlag,
			signModeFlag,
			signMemoFlag,
			mpcServerFlag,
			mpcKeystoreFlag,
			mpcPasswordFlag,
			signTypeFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
			signTimeoutFlag,
		},
	}
)

func signPlainText(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	mpcCfg.NeedKeyStore = true
	err = checkAndInitMpcConfig(ctx, true)
	if err != nil {
		return err
	}

	var signContent string
	var msgContext []string

	if signMessageArg != "" {
		signContent = signMessageArg
		msgContext = []string{"hexstring", signMessageArg}
	} else {
		signContent = msgHashArg
		msgContext = []string{"plaintext", msgContextArg}

		hash := crypto.Keccak256Hash([]byte(msgContextArg))
		if hash != common.HexToHash(msgHashArg) {
			return errors.New("message hash is not the keccak256 hash of plaintext msgContext")
		}
	}

	if signMemoArg != "" {
		msgContext = append(msgContext, signMemoArg)
	}

	keyID, rsvs, err := mpcrpc.DoSign(mpcPublicKey, []string{signContent}, msgContext)
	if err != nil {
		log.Error("mpc sign failed", "err", err)
		return err
	}
	log.Info("mpc sign success", "keyID", keyID)

	if len(rsvs) != 1 {
		log.Error("mpc sign result rsv count is wrong", "have", len(rsvs), "want", 1)
		return errors.New("mpc sign result rsv count is wrong")
	}
	rsv := rsvs[0]

	signature := common.FromHex(rsv)
	if len(signature) != crypto.SignatureLength {
		log.Error("mpc sign result rsv length is wrong", "rsv", rsv)
		return errors.New("mpc sign result rsv length is wrong")
	}

	fmt.Println("rsv is", rsv)
	return nil
}
