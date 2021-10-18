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
	signStatus, err := mpcrpc.GetSignStatus(keyID, mpcCfg.RPCAddress)
	if err != nil {
		return err
	}

	jsData, err := json.MarshalIndent(signStatus, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsData))
	return nil
}
