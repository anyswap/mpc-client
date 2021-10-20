package main

import (
	"encoding/json"
	"fmt"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/urfave/cli/v2"
)

var (
	getSignStatusCommand = &cli.Command{
		Action:      getSignStatus,
		Name:        "getsignstatus",
		Usage:       "get sign status",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			keyIDFlag,
			mpcServerFlag,
			mpcDKGFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func getSignStatus(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	keyID := ctx.String(keyIDFlag.Name)
	isDKG := ctx.Bool(mpcDKGFlag.Name)
	if isDKG {
		dkgStatus, err := mpcrpc.GetReqAddrStatus(keyID, mpcCfg.RPCAddress)
		if err != nil {
			return err
		}

		jsData, err := json.MarshalIndent(dkgStatus, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsData))
	} else {
		signStatus, err := mpcrpc.GetSignStatus(keyID, mpcCfg.RPCAddress)
		if err != nil {
			return err
		}

		jsData, err := json.MarshalIndent(signStatus, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsData))
	}
	return nil
}
