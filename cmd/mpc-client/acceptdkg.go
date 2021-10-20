package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc"
	"github.com/urfave/cli/v2"
)

func acceptDKG(ctx *cli.Context) (err error) {
	keyID := ctx.String(keyIDFlag.Name)
	interactiveMode := !ctx.Bool(nonInteractiveFlag.Name)
	if !interactiveMode {
		isAgree := ctx.Bool(agreeSignFlag.Name) && !ctx.Bool(disagreeSignFlag.Name) // disagree first
		agreeResult := getAgreeResult(isAgree)
		return doAcceptDKGNoninteractively(keyID, agreeResult)
	}

	_, err = getDKGInfoByKeyID(keyID)
	if err != nil {
		return err
	}

	isAgree := askForReply("Do you agree this dkg?")
	agreeResult := getAgreeResult(isAgree)
	return doAcceptDKG(keyID, agreeResult)
}

func getDKGInfoByKeyID(keyID string) (dkgInfo *mpcrpc.ReqAddrInfoData, err error) {
	dkgInfos, err := mpcrpc.GetCurNodeReqAddrInfo()
	if err != nil {
		log.Error("getCurNodeReqAddrInfo failed", "err", err)
		return nil, err
	}

	for _, info := range dkgInfos {
		if info != nil && strings.EqualFold(info.Key, keyID) {
			dkgInfo = info
			break
		}
	}

	if dkgInfo == nil {
		return nil, errors.New("keyID is not found in dkg accept list")
	}

	jsData, err := json.MarshalIndent(dkgInfo, "", "  ")
	if err != nil {
		return nil, err
	}
	fmt.Println("dkg info is", string(jsData))

	return dkgInfo, nil
}

func doAcceptDKGNoninteractively(keyID, agreeResult string) (err error) {
	if !isAllKeyID(keyID) {
		dkgInfo, errt := getDKGInfoByKeyID(keyID)
		if errt != nil {
			return errt
		}
		log.Info("get dkg info success", "cointype", dkgInfo.Cointype, "account", dkgInfo.Account)
		return doAcceptDKG(keyID, agreeResult)
	}

	dkgInfos, err := mpcrpc.GetCurNodeReqAddrInfo()
	if err != nil {
		log.Error("getCurNodeReqAddrInfo failed", "err", err)
		return err
	}

	for _, info := range dkgInfos {
		go func(dkgInfo *mpcrpc.ReqAddrInfoData) {
			errt := doAcceptDKG(dkgInfo.Key, agreeResult)
			if errt != nil {
				log.Warn("accept dkg failed", "dkgInfo", dkgInfo, "agreeResult", agreeResult, "err", err)
			}
		}(info)
	}
	return nil
}

func doAcceptDKG(keyID, agreeResult string) (err error) {
	result, err := mpcrpc.DoAcceptReqAddr(keyID, agreeResult)
	if err != nil {
		log.Error("mpc accept dkg failed", "keyID", keyID, "rpcResult", result, "err", err)
		return err
	}
	log.Info("mpc accept dkg finished", "keyID", keyID, "agreeResult", agreeResult, "rpcResult", result)
	return nil
}
