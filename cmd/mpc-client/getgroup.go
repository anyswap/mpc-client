package main

import (
	"encoding/json"
	"fmt"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/urfave/cli/v2"
)

var (
	getGroupCommand = &cli.Command{
		Action:      getGroup,
		Name:        "getgroup",
		Usage:       "get group info",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			groupIDFlag,
			mpcServerFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func getGroup(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	groupID := ctx.String(groupIDFlag.Name)
	groupInfo, err := mpcrpc.GetGroupByID(groupID, mpcCfg.RPCAddress)
	if err != nil {
		return err
	}

	jsData, err := json.MarshalIndent(groupInfo, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsData))
	return nil
}
