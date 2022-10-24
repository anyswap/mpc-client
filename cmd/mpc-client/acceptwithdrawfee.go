package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var (
	acceptWithdrawFeeCommand = &cli.Command{
		Action:      acceptWithdrawFee,
		Name:        "acceptwithdrawfee",
		Usage:       "start accept withdraw fee in batch",
		ArgsUsage:   "",
		Description: ``,
		Flags: []cli.Flag{
			senderAddrFlag,
			receiversAddrFlag,
			multicallsAddrFlag,
			mpcServerFlag,
			mpcKeystoreFlag,
			mpcPasswordFlag,
			apiPrefixFlag,
			rpcTimeoutFlag,
		},
	}

	feeAllowedSignSender  string
	feeAllowedReceivers   []string
	allowedMulticallAddrs []string
)

func acceptWithdrawFee(ctx *cli.Context) (err error) {
	utils.SetLogger(ctx)
	mpcCfg.NeedKeyStore = true
	err = checkAndInitMpcConfig(ctx, false)
	if err != nil {
		return err
	}

	feeAllowedSignSender = ctx.String(senderAddrFlag.Name)
	if feeAllowedSignSender == "" {
		return errors.New("must specify withdraw fee sender (with --from option)")
	}
	log.Infof("withdraw fee allowed sign sender is %v", feeAllowedSignSender)

	receiverArg := ctx.String(receiversAddrFlag.Name)
	if receiverArg != "" {
		feeAllowedReceivers = strings.Split(receiverArg, ",")
	}
	if len(feeAllowedReceivers) == 0 {
		return errors.New("must specify allowed receivers (with --receivers option)")
	}
	log.Infof("withdraw fee allowed receivers are %v", feeAllowedReceivers)

	multicallArg := ctx.String(multicallsAddrFlag.Name)
	if multicallArg != "" {
		allowedMulticallAddrs = strings.Split(multicallArg, ",")
	}
	log.Infof("withdraw fee allowed multicall contracts are %v", allowedMulticallAddrs)

	var loop uint64
	for {
		loop++
		log.Infof("start accept loop %v", loop)

		signInfos, errf := mpcrpc.GetCurNodeSignInfo(0)
		if errf != nil {
			log.Error("getCurNodeSignInfo failed", "err", errf)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, info := range signInfos {
			keyID := info.Key

			isAgree, isIgnore, errf := verifyWithdrawFeeSignInfo(info)
			if isIgnore {
				log.Debug("ignore sign info", "keyID", keyID, "err", errf)
				continue
			}
			if !isAgree {
				log.Warn("diagree sign info", "keyID", keyID, "err", errf)
			}

			agreeResult := getAgreeResult(isAgree)
			errf = doAcceptSign(keyID, agreeResult, info.MsgHash, info.MsgContext)
			if errf != nil {
				log.Warn("call accept sign error", "keyID", keyID, "err", errf)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func verifyWithdrawFeeSignInfo(info *mpcrpc.SignInfoData) (isAgree, isIgnore bool, err error) {
	msgHashes := info.MsgHash
	msgContexts := info.MsgContext

	isIgnore = true

	if len(msgHashes) != 1 {
		return isAgree, isIgnore, errors.New("mismatch message hash length")
	}

	if len(msgContexts) != 4 {
		return isAgree, isIgnore, errors.New("mismatch message context length")
	}

	msgHash := msgHashes[0]
	msgContextType := msgContexts[0]
	msgContext := msgContexts[1]
	chainIDStr := msgContexts[2]
	signatureHex := msgContexts[3]

	// verify message context
	if strings.ToLower(msgContextType) != "withdrawfee" {
		return isAgree, isIgnore, errors.New("mismatch message context type")
	}

	chainID, ok := new(big.Int).SetString(chainIDStr, 0)
	if !ok {
		return isAgree, isIgnore, fmt.Errorf("wrong chainID '%v'", chainIDStr)
	}

	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return isAgree, isIgnore, fmt.Errorf("wrong signature format: %w", err)
	}

	recoveredPub, err := crypto.Ecrecover(common.HexToHash(msgHash).Bytes(), signature)
	if err != nil {
		return isAgree, isIgnore, fmt.Errorf("recover signature failed: %w", err)
	}
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if !strings.EqualFold(feeAllowedSignSender, recoveredAddr.String()) {
		return isAgree, isIgnore, fmt.Errorf("mismatch signature signer: %v", recoveredAddr.String())
	}

	var rawTx types.Transaction
	err = json.Unmarshal([]byte(msgContext), &rawTx)
	if err != nil {
		return isAgree, isIgnore, fmt.Errorf("json unmarshal msgContext failed. %w", err)
	}

	log.Printf("the sign is sending the following tx to block chain (chainID: %v)", chainIDStr)
	if errf := printTx(&rawTx, true); errf != nil {
		log.Warn("print transaction failed", "err", errf)
	}

	err = verifyWithdrawFeeTx(&rawTx)
	if err != nil {
		return isAgree, isIgnore, err
	}

	// the sign info message context is right, will not ignore it from now on
	isIgnore = false

	chainSigner := types.NewEIP155Signer(chainID)
	calcedHash := chainSigner.Hash(&rawTx)
	err = checkMessageHash(calcedHash, msgHash)
	if err != nil {
		return isAgree, isIgnore, err
	}

	return true, isIgnore, nil
}

func verifyWithdrawFeeTx(tx *types.Transaction) error {
	to := tx.To()
	if to == nil {
		return errors.New("create contract tx not allowed")
	}

	txVal := tx.Value()
	if txVal != nil && txVal.Sign() > 0 { // send native coin tx
		receiver := to.String()
		return checkFeeReceiver(receiver)
	}

	txData := tx.Data()
	if len(txData) < 4 {
		return errors.New("input data is too short")
	}

	if isAllowedMulticallAddr(to.String()) {
		return checkMulticallData(tx, txData)
	}

	funcHash := hex.EncodeToString(txData[:4])
	data := txData[4:]

	switch funcHash {
	case "87cc6e2f": // "anySwapFeeTo(address,uint256)"
	case "ada82c7d": // "withdrawAccruedFees()"
	case "a9059cbb": // "transfer(address,uint256)"
		if len(data) != 64 {
			return errors.New("wrong transfer tx input data length")
		}
		receiver := common.BytesToAddress(data[:32]).String()
		return checkFeeReceiver(receiver)
	default:
		return errors.New("function hash not allowed")
	}

	return nil
}

func isAllowedMulticallAddr(address string) bool {
	for _, addr := range allowedMulticallAddrs {
		if strings.EqualFold(addr, address) {
			return true
		}
	}
	return false
}

func checkFeeReceiver(receiver string) error {
	for _, addr := range feeAllowedReceivers {
		if strings.EqualFold(addr, receiver) {
			return nil
		}
	}
	return errors.New("mismatch receivers")
}

func checkMulticallData(tx *types.Transaction, data []byte) error {
	var callArgs []multicallArg
	var err error

	funcHash := hex.EncodeToString(data[:4])
	data = data[4:]

	switch funcHash {
	case "252dba42": // "aggregate((address,bytes)[])"
		callArgs, err = decodeMulticallArg1(data)
	case "82ad56cb": // "aggregate3((address,bool,bytes)[])"
		callArgs, err = decodeMulticallArg2(data)
	case "174dea71": // "aggregate3Value((address,bool,uint256,bytes)[])"
		callArgs, err = decodeMulticallArg3(data)
	case "c3077fa9": // "blockAndAggregate((address,bytes)[])"
		callArgs, err = decodeMulticallArg1(data)
	case "bce38bd7": // "tryAggregate(bool,(address,bytes)[])"
		callArgs, err = decodeMulticallArg1(data[32:])
	case "399542e9": // "tryBlockAndAggregate(bool,(address,bytes)[])"
		callArgs, err = decodeMulticallArg1(data[32:])
	default:
		return errors.New("multicall: function hash not allowed")
	}
	if err != nil {
		return fmt.Errorf("multicall: decode data failed: %w", err)
	}

	for _, callArg := range callArgs {
		err = checkMulticallArg(tx, callArg)
		if err != nil {
			return err
		}
	}
	return nil
}

type multicallArg struct {
	target       string
	allowFailure bool
	value        *big.Int
	callData     []byte
}

func checkMulticallArg(tx *types.Transaction, callArg multicallArg) error {
	if callArg.value != nil && callArg.value.Sign() > 0 {
		return checkFeeReceiver(callArg.target)
	}

	callData := callArg.callData
	if len(callData) < 4 {
		return fmt.Errorf("multicall: call data is too short, call arg is %v", callArg)
	}

	funcHash := hex.EncodeToString(callData[:4])
	data := callData[4:]

	switch funcHash {
	case "095ea7b3": // "approve(address,uint256)"
		if len(data) != 64 {
			return errors.New("multicall: wrong approve tx input data length")
		}
		approveTo := common.BytesToAddress(data[:32])
		if approveTo != *tx.To() {
			return errors.New("multicall: approve to other address")
		}
	case "23b872dd": // "transferFrom(address,address,uint256)",
		if len(data) != 96 {
			return errors.New("multicall: wrong transfer from tx input data length")
		}
		receiver := common.BytesToAddress(data[32:64]).String()
		return checkFeeReceiver(receiver)
	default:
		return errors.New("multicall: function hash not allowed")
	}
	return nil
}

// (address,bytes)[]
func decodeMulticallArg1(data []byte) ([]multicallArg, error) {
	var callArgs []multicallArg
	count := new(big.Int).SetBytes(getData(data, 32, 32)).Uint64()
	data = data[64:]
	for i := uint64(0); i < count; i++ {
		offset := new(big.Int).SetBytes(getData(data, i*32, 32)).Uint64()
		idata := data[offset:]

		var callArg multicallArg
		callArg.target = common.BytesToAddress(getData(idata, 0, 32)).String()
		callDataLen := new(big.Int).SetBytes(getData(idata, 64, 32)).Uint64()
		callArg.callData = idata[96 : 96+callDataLen]
		callArgs = append(callArgs, callArg)
	}
	return callArgs, nil
}

// (address,bool,bytes)[]
func decodeMulticallArg2(data []byte) ([]multicallArg, error) {
	var callArgs []multicallArg
	count := new(big.Int).SetBytes(getData(data, 32, 32)).Uint64()
	data = data[64:]
	for i := uint64(0); i < count; i++ {
		offset := new(big.Int).SetBytes(getData(data, i*32, 32)).Uint64()
		idata := data[offset:]

		var callArg multicallArg
		callArg.target = common.BytesToAddress(getData(idata, 0, 32)).String()
		callArg.allowFailure = new(big.Int).SetBytes(getData(idata, 32, 32)).Sign() == 0
		callDataLen := new(big.Int).SetBytes(getData(idata, 96, 32)).Uint64()
		callArg.callData = idata[128 : 128+callDataLen]
		callArgs = append(callArgs, callArg)
	}
	return callArgs, nil
}

// (address,bool,uint256,bytes)[]
func decodeMulticallArg3(data []byte) ([]multicallArg, error) {
	var callArgs []multicallArg
	count := new(big.Int).SetBytes(getData(data, 32, 32)).Uint64()
	data = data[64:]
	for i := uint64(0); i < count; i++ {
		offset := new(big.Int).SetBytes(getData(data, i*32, 32)).Uint64()
		idata := data[offset:]

		var callArg multicallArg
		callArg.target = common.BytesToAddress(getData(idata, 0, 32)).String()
		callArg.allowFailure = new(big.Int).SetBytes(getData(idata, 32, 32)).Sign() == 0
		callArg.value = new(big.Int).SetBytes(getData(idata, 64, 32))
		callDataLen := new(big.Int).SetBytes(getData(idata, 128, 32)).Uint64()
		callArg.callData = idata[160 : 160+callDataLen]
		callArgs = append(callArgs, callArg)
	}
	return callArgs, nil
}

// getData returns a slice from the data based on the start and size and pads
// up to size with zero's. This function is overflow safe.
func getData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return common.RightPadBytes(data[start:end], int(size))
}
