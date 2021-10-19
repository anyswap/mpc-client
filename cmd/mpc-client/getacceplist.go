package main

import (
	"encoding/json"
	"fmt"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/urfave/cli/v2"
)

var (
	getAcceptListCommand = &cli.Command{
		Action:      getAcceptList,
		Name:        "getacceptlist",
		Usage:       "get accept list",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			mpcUserFlag,
			mpcServerFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func getAcceptList(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	user := ctx.String(mpcUserFlag.Name)
	accpetList, err := mpcrpc.GetAcceptList(user)
	if err != nil {
		return err
	}
	jsData, err := json.MarshalIndent(accpetList, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsData))
	fmt.Println("accept list length is", len(accpetList))
	return nil
}
