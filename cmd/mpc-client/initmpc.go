package main

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
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
		signMode := ctx.Uint64(signModeFlag.Name)
		mpcCfg.Mode = &signMode

		if !mpcCfg.IsDKG {
			mpcPublicKey = ctx.String(pubkeyFlag.Name)
			if mpcPublicKey == "" {
				return errors.New("empty mpc public key")
			}
			msgHashArg = ctx.String(msgHashFlag.Name)
			if msgHashArg != "" && !strings.EqualFold(common.HexToHash(msgHashArg).String(), msgHashArg) {
				return errors.New("wrong message hash argument")
			}
			msgContextArg = ctx.String(msgContextFlag.Name)
		}
		signMemoArg = ctx.String(signMemoFlag.Name)
	}

	mergeConfigFromConfigFile(ctx)

	mpcrpc.Init(&mpcCfg, isSign)
	return nil
}

func mergeConfigFromConfigFile(ctx *cli.Context) {
	configFile := utils.GetConfigFilePath(ctx)
	if configFile == "" {
		return
	}
	config := loadConfigFile(configFile)
	if config == nil {
		return
	}

	if config.MPC.APIPrefix != "" && !ctx.IsSet(apiPrefixFlag.Name) {
		mpcCfg.APIPrefix = config.MPC.APIPrefix
	}
	if config.MPC.RPCAddress != "" && !ctx.IsSet(mpcServerFlag.Name) {
		mpcCfg.RPCAddress = config.MPC.RPCAddress
	}
	if config.MPC.RPCTimeout != 0 && !ctx.IsSet(rpcTimeoutFlag.Name) {
		mpcCfg.RPCTimeout = config.MPC.RPCTimeout
	}
	if config.MPC.KeystoreFile != "" && !ctx.IsSet(mpcKeystoreFlag.Name) {
		mpcCfg.KeystoreFile = config.MPC.KeystoreFile
	}
	if config.MPC.PasswordFile != "" && !ctx.IsSet(mpcPasswordFlag.Name) {
		mpcCfg.PasswordFile = config.MPC.PasswordFile
	}
	if config.MPC.SignTimeout != 0 && !ctx.IsSet(signTimeoutFlag.Name) {
		mpcCfg.SignTimeout = config.MPC.SignTimeout
	}
	if config.MPC.SignType != "" && !ctx.IsSet(signTypeFlag.Name) {
		mpcCfg.SignType = config.MPC.SignType
	}
	if config.MPC.SignGroup != "" && !ctx.IsSet(gidFlag.Name) {
		mpcCfg.SignGroup = config.MPC.SignGroup
	}
	if config.MPC.Threshold != "" && !ctx.IsSet(thresholdFlag.Name) {
		mpcCfg.Threshold = config.MPC.Threshold
	}
	if config.MPC.Mode != nil && !ctx.IsSet(signModeFlag.Name) {
		mpcCfg.Mode = config.MPC.Mode
	}
}

// Config toml config
type Config struct {
	MPC *mpcrpc.MPCConfig
}

func loadConfigFile(configFile string) (config *Config) {
	log.Info("load config file", "configFile", configFile)
	if !common.FileExist(configFile) {
		log.Fatalf("config file '%v' not exist", configFile)
	}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalf("load config file error (toml DecodeFile): %v", err)
	}

	var bs []byte
	if log.JSONFormat {
		bs, _ = json.Marshal(config)
	} else {
		bs, _ = json.MarshalIndent(config, "", "  ")
	}
	log.Println("load config file finished.", string(bs))
	return config
}
