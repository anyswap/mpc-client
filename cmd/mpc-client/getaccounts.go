package main

import (
	"encoding/json"
	"fmt"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/urfave/cli/v2"
)

var (
	getAccountsCommand = &cli.Command{
		Action:      getAccounts,
		Name:        "getaccounts",
		Usage:       "get accounts info",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			mpcUserFlag,
			signModeFlag,
			mpcServerFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func getAccounts(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	user := ctx.String(mpcUserFlag.Name)
	accoutsInfo, err := mpcrpc.GetAccounts(user)
	if err != nil {
		return err
	}

	jsData, err := json.MarshalIndent(accoutsInfo, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsData))
	return nil
}
