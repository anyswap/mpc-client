package main

import (
	"fmt"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/urfave/cli/v2"
)

var (
	getEnodeCommand = &cli.Command{
		Action:      getEnode,
		Name:        "getenode",
		Usage:       "get enode info",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			mpcServerFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func getEnode(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	enode, err := mpcrpc.GetEnode(mpcCfg.RPCAddress)
	if err != nil {
		return err
	}
	fmt.Println("enode is ", enode)
	return nil
}
