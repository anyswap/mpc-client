package main

import (
	"fmt"
	"strings"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
			showEnodeSigFlag,
			mpcServerFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}
)

func getEnode(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	showEnodeSig := ctx.Bool(showEnodeSigFlag.Name)
	if showEnodeSig {
		mpcCfg.NeedKeyStore = true
	}
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	enode, err := mpcrpc.GetEnode(mpcCfg.RPCAddress)
	if err != nil {
		return err
	}
	fmt.Println("enode is", enode)

	if !showEnodeSig {
		return nil
	}

	startIndex := strings.Index(enode, "enode://")
	endIndex := strings.Index(enode, "@")
	if startIndex == -1 || endIndex == -1 {
		return fmt.Errorf("wrong enode '%v'", enode)
	}
	enodePubkey := enode[startIndex+8 : endIndex]
	sig, err := mpcrpc.SignContent([]byte(enodePubkey))
	if err != nil {
		return err
	}
	fmt.Println("enode sig is", hexutil.Encode(sig))
	return nil
}
