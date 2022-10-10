package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var (
	withdrawFeeCommand = &cli.Command{
		Action:      withdrawFee,
		Name:        "withdrawfee",
		Usage:       "send withdraw fee tx",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			pubkeyFlag,
			gidFlag,
			thresholdFlag,
			signModeFlag,
			signMemoFlag,
			mpcServerFlag,
			mpcKeystoreFlag,
			mpcPasswordFlag,
			signTypeFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
			signTimeoutFlag,
			gatewaysFlag,
			chainIDFlag,
			fromAddrFlag,
			toAddrFlag,
			nonceFlag,
			valueFlag,
			gasLimitFlag,
			gasPriceFlag,
			inputFlag,
			dryrunFlag,
		},
	}
)

func checkWithdrawFeeArguments(ctx *cli.Context) (err error) {
	txArgs.gateways = ctx.StringSlice(gatewaysFlag.Name)
	txArgs.gasLimit = ctx.Uint64(gasLimitFlag.Name)
	txArgs.dryrun = ctx.Bool(dryrunFlag.Name)

	fromAddrStr := ctx.String(fromAddrFlag.Name)
	if !common.IsHexAddress(fromAddrStr) {
		return fmt.Errorf("wrong from address %v", fromAddrStr)
	}
	txArgs.from = common.HexToAddress(fromAddrStr)

	toAddrStr := ctx.String(toAddrFlag.Name)
	if !common.IsHexAddress(toAddrStr) {
		return fmt.Errorf("wrong to address %v", toAddrStr)
	}
	txArgs.to = common.HexToAddress(toAddrStr)

	var ok bool
	gasPriceStr := ctx.String(gasPriceFlag.Name)
	txArgs.gasPrice, ok = new(big.Int).SetString(gasPriceStr, 0)
	if !ok {
		return fmt.Errorf("wrong gas price %v", gasPriceStr)
	}

	nodeChainIDStr := ctx.String(chainIDFlag.Name)
	txArgs.chainID, ok = new(big.Int).SetString(nodeChainIDStr, 0)
	if !ok {
		return fmt.Errorf("wrong chain Id %v", nodeChainIDStr)
	}

	accNonceStr := ctx.String(nonceFlag.Name)
	if accNonceStr != "" {
		txArgs.accNonce, ok = new(big.Int).SetString(accNonceStr, 0)
		if !ok {
			return fmt.Errorf("wrong account nonce %v", accNonceStr)
		}
	}

	valueStr := ctx.String(valueFlag.Name)
	if valueStr != "" {
		txArgs.value, ok = new(big.Int).SetString(valueStr, 0)
		if !ok {
			return fmt.Errorf("wrong value %v", valueStr)
		}
	}

	inputData := ctx.String(inputFlag.Name)
	if inputData != "" {
		txArgs.input, err = hexutil.Decode(inputData)
		if err != nil {
			return fmt.Errorf("wrong input data %v, err=%v", inputData, err)
		}
	}

	log.Info("check arguments pass")
	return nil
}

func withdrawFee(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	mpcCfg.NeedKeyStore = true
	err = checkAndInitMpcConfig(ctx, true)
	if err != nil {
		return err
	}
	err = checkWithdrawFeeArguments(ctx)
	if err != nil {
		return err
	}

	err = dailGateways(txArgs.gateways)
	if err != nil {
		return err
	}

	var nonce uint64
	if txArgs.accNonce != nil {
		nonce = txArgs.accNonce.Uint64()
	} else {
		nonce, err = getPendingNonce(txArgs.from)
		if err != nil {
			log.Error("get account nonce failed", "account", txArgs.from.String(), "err", err)
			return err
		}
		log.Info("get account nonce success", "account", txArgs.from.String(), "nonce", nonce)
	}

	var rawTx *types.Transaction
	if txArgs.createContract {
		rawTx = types.NewContractCreation(nonce, txArgs.value, txArgs.gasLimit, txArgs.gasPrice, txArgs.input)
	} else {
		rawTx = types.NewTransaction(nonce, txArgs.to, txArgs.value, txArgs.gasLimit, txArgs.gasPrice, txArgs.input)
	}
	log.Info("create raw tx success")
	_ = printTx(rawTx, true)

	chainSigner := types.NewEIP155Signer(txArgs.chainID)
	msgHash := chainSigner.Hash(rawTx)
	txJSON, err := json.Marshal(rawTx)
	if err != nil {
		log.Error("json marshal tx failed")
		return err
	}

	senderSignature, err := mpcrpc.SignWithKey(msgHash[:])
	if err != nil {
		return err
	}

	msgContext := []string{"withdrawfee", string(txJSON), txArgs.chainID.String(), hex.EncodeToString(senderSignature)}
	if signMemoArg != "" {
		msgContext = append(msgContext, signMemoArg)
	}

	keyID, rsvs, err := mpcrpc.DoSign(mpcPublicKey, []string{msgHash.String()}, msgContext)
	if err != nil {
		log.Error("mpc sign failed", "err", err)
		return err
	}
	log.Info("mpc sign success", "keyID", keyID)

	if len(rsvs) != 1 {
		log.Error("mpc sign result rsv count is wrong", "have", len(rsvs), "want", 1)
		return errors.New("mpc sign result rsv count is wrong")
	}
	rsv := rsvs[0]

	signature := common.FromHex(rsv)
	if len(signature) != crypto.SignatureLength {
		log.Error("mpc sign result rsv length is wrong", "rsv", rsv)
		return errors.New("mpc sign result rsv length is wrong")
	}

	signedTx, err := rawTx.WithSignature(chainSigner, signature)
	if err != nil {
		log.Error("sign tx failed", "err", err)
		return err
	}

	sender, err := types.Sender(chainSigner, signedTx)
	if err != nil {
		log.Error("get sender from signed tx failed", "err", err)
		return err
	}

	if sender != txArgs.from {
		log.Error("sender mismatch", "signer", sender.String(), "sender", txArgs.from.String())
		return errors.New("sender mismatch")
	}

	txHash := signedTx.Hash().String()

	log.Info("mpc sign tx success", "txHash", txHash, "sender", sender.String())
	_ = printTx(signedTx, false)

	if !txArgs.dryrun {
		err = sendSignedTransaction(signedTx)
		if err != nil {
			log.Error("send tx failed", "err", err)
			return err
		}
		log.Info("send tx success", "txHash", txHash)
	}
	return nil
}
