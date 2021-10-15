package main

import (
	"errors"
	"strings"

	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

var (
	mpcPublicKey string

	msgHashArg    string
	msgContextArg string
	signMemoArg   string

	mpcCfg mpcrpc.MPCConfig
)

func checkAndInitMpcConfig(ctx *cli.Context, isSign bool) (err error) {
	mpcCfg.APIPrefix = ctx.String(apiPrefixFlag.Name)
	mpcCfg.RPCAddress = ctx.String(mpcServerFlag.Name)
	mpcCfg.RPCTimeout = ctx.Uint64(rpcTimeoutFlag.Name)
	mpcCfg.KeystoreFile = ctx.String(mpcKeystoreFlag.Name)
	mpcCfg.PasswordFile = ctx.String(mpcPasswordFlag.Name)

	if isSign {
		mpcCfg.SignTimeout = ctx.Uint64(signTimeoutFlag.Name)
		mpcCfg.SignType = ctx.String(signTypeFlag.Name)
		mpcCfg.SignGroup = ctx.String(gidFlag.Name)
		mpcCfg.Threshold = ctx.String(thresholdFlag.Name)
		mpcCfg.Mode = ctx.Uint64(signModeFlag.Name)

		mpcPublicKey = ctx.String(pubkeyFlag.Name)
		if mpcPublicKey == "" {
			return errors.New("empty mpc public key")
		}
		msgHashArg = ctx.String(msgHashFlag.Name)
		if msgHashArg != "" && !strings.EqualFold(common.HexToHash(msgHashArg).String(), msgHashArg) {
			return errors.New("wrong message hash argument")
		}
		msgContextArg = ctx.String(msgContextFlag.Name)
		signMemoArg = ctx.String(signMemoFlag.Name)
	}

	mpcrpc.Init(&mpcCfg, isSign)
	return nil
}
