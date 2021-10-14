package main

import (
	"github.com/urfave/cli/v2"
)

var (
	pubkeyFlag = &cli.StringFlag{
		Name:  "pubkey",
		Usage: "mpc public key",
	}
	msgHashFlag = &cli.StringFlag{
		Name:  "msghash",
		Usage: "mpc sign message hash",
	}
	msgContextFlag = &cli.StringFlag{
		Name:  "msgcontext",
		Usage: "mpc sign message context",
	}
	keyIDFlag = &cli.StringFlag{
		Name:  "key",
		Usage: "mpc sign key ID",
	}
	apiPrefixFlag = &cli.StringFlag{
		Name:  "apiPrefix",
		Usage: "mpc rpc apiPrefix",
		Value: "dcrm_",
	}
	rpcTimeoutFlag = &cli.Uint64Flag{
		Name:  "rpcTimeout",
		Usage: "mpc rpc timeout of seconds",
		Value: 20,
	}
	signTimeoutFlag = &cli.Uint64Flag{
		Name:  "signTimeout",
		Usage: "mpc sign timeout of seconds",
		Value: 120,
	}
	signTypeFlag = &cli.StringFlag{
		Name:  "keytype",
		Usage: "mpc sign algorithm type",
		Value: "ECDSA",
	}
	gidFlag = &cli.StringFlag{
		Name:  "gid",
		Usage: "mpc sign group ID",
	}
	thresholdFlag = &cli.StringFlag{
		Name:  "ts",
		Usage: "mpc sign threshold",
		Value: "3/5",
	}
	signModeFlag = &cli.Uint64Flag{
		Name:  "mode",
		Usage: "mpc sign mode (private=1/managed=0)",
		Value: 0,
	}
	mpcServerFlag = &cli.StringFlag{
		Name:  "url",
		Usage: "mpc server URL",
	}
	mpcKeystoreFlag = &cli.StringFlag{
		Name:  "keystore",
		Usage: "mpc user keystore file",
	}
	mpcPasswordFlag = &cli.StringFlag{
		Name:  "passwd",
		Usage: "mpc user password file",
	}

	gatewayFlag = &cli.StringFlag{
		Name:  "gateway",
		Usage: "gateway URL of blockchain full node",
	}
	chainIDFlag = &cli.StringFlag{
		Name:  "chainID",
		Usage: "blockchain ID",
	}
	fromAddrFlag = &cli.StringFlag{
		Name:  "from",
		Usage: "tx sender address",
	}
	toAddrFlag = &cli.StringFlag{
		Name:  "to",
		Usage: "tx receiver address",
	}
	inputFlag = &cli.StringFlag{
		Name:  "input",
		Usage: "tx input data",
	}
	gasLimitFlag = &cli.Uint64Flag{
		Name:  "gas",
		Usage: "tx gas limit",
		Value: 90000,
	}
	gasPriceFlag = &cli.StringFlag{
		Name:  "gasPrice",
		Usage: "tx gas price in Wei",
	}
	nonceFlag = &cli.StringFlag{
		Name:  "nonce",
		Usage: "tx nonce",
	}
	valueFlag = &cli.StringFlag{
		Name:  "value",
		Usage: "tx value of native coins",
	}
	dryrunFlag = &cli.BoolFlag{
		Name:  "dryrun",
		Usage: "dry run",
	}
)
