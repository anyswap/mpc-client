package main

import (
	"fmt"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

var (
	doDKGCommand = &cli.Command{
		Action:      doDKG,
		Name:        "dkg",
		Usage:       "generate public key",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			gidFlag,
			thresholdFlag,
			signModeFlag,
			enodeSigsFlag,
			mpcServerFlag,
			mpcKeystoreFlag,
			mpcPasswordFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
			signTimeoutFlag,
		},
	}
)

func doDKG(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	mpcCfg.NeedKeyStore = true
	mpcCfg.IsDKG = true
	err = checkAndInitMpcConfig(ctx, true)
	if err != nil {
		return err
	}

	enodeSigs := ctx.StringSlice(enodeSigsFlag.Name)
	keyID, pubkey, err := mpcrpc.DoDKG(enodeSigs)
	if err != nil {
		log.Error("mpc dkg failed", "err", err)
		return err
	}
	log.Info("mpc dkg success", "keyID", keyID)

	pkBytes := common.FromHex(pubkey)
	if len(pkBytes) != 65 || pkBytes[0] != 4 {
		return fmt.Errorf("wrong mpc public key '%v'", pubkey)
	}

	fmt.Println("pubkey is", pubkey)
	return nil
}
