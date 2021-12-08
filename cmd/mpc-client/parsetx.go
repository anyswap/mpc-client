package main

import (
	"encoding/hex"

	"github.com/anyswap/mpc-client/log"
	"github.com/ethereum/go-ethereum/core/types"
)

var knownContractMethods = map[string]string{
	// ======= AnyswapV5ERC20 ===================
	"3644e515": "DOMAIN_SEPARATOR()",
	"30adf81f": "PERMIT_TYPEHASH()",
	"ec126c77": "Swapin(bytes32,address,uint256)",
	"628d6cba": "Swapout(uint256,address)",
	"00bf26f4": "TRANSFER_TYPEHASH()",
	"dd62ed3e": "allowance(address,address)",
	"0d707df8": "applyMinter()",
	"d93f2445": "applyVault()",
	"095ea7b3": "approve(address,uint256)",
	"cae9ca51": "approveAndCall(address,uint256,bytes)",
	"70a08231": "balanceOf(address)",
	"9dc29fac": "burn(address,uint256)",
	"5f9b105d": "changeMPCOwner(address)",
	"60e232a9": "changeVault(address)",
	"313ce567": "decimals()",
	"6a42b8f8": "delay()",
	"a29dff72": "delayDelay()",
	"c3081240": "delayMinter()",
	"87689e28": "delayVault()",
	"d0e30db0": "deposit()",
	"b6b55f25": "deposit(uint256)",
	"6e553f65": "deposit(uint256,address)",
	"bebbf4d0": "depositVault(uint256,address)",
	"81a37c18": "depositWithPermit(address,uint256,uint256,uint8,bytes32,bytes32,address)",
	"f954734e": "depositWithTransferPermit(address,uint256,uint256,uint8,bytes32,bytes32,address)",
	"a045442c": "getAllMinters()",
	"2ebe3fbb": "initVault(address)",
	"aa271e1a": "isMinter(address)",
	"40c10f19": "mint(address,uint256)",
	"8623ec7b": "minters(uint256)",
	"f75c2664": "mpc()",
	"06fdde03": "name()",
	"7ecebe00": "nonces(address)",
	"8da5cb5b": "owner()",
	"4ca8f0ed": "pendingDelay()",
	"91c5df49": "pendingMinter()",
	"52113ba7": "pendingVault()",
	"d505accf": "permit(address,address,uint256,uint256,uint8,bytes32,bytes32)",
	"cfbd4885": "revokeMinter(address)",
	"fca3b5aa": "setMinter(address)",
	"6817031b": "setVault(address)",
	"c4b740f5": "setVaultOnly(bool)",
	"95d89b41": "symbol()",
	"18160ddd": "totalSupply()",
	"a9059cbb": "transfer(address,uint256)",
	"4000aea0": "transferAndCall(address,uint256,bytes)",
	"23b872dd": "transferFrom(address,address,uint256)",
	"605629d6": "transferWithPermit(address,address,uint256,uint256,uint8,bytes32,bytes32)",
	"6f307dc3": "underlying()",
	"fbfa77cf": "vault()",
	"3ccfd60b": "withdraw()",
	"2e1a7d4d": "withdraw(uint256)",
	"00f714ce": "withdraw(uint256,address)",
	"0039d6ec": "withdrawVault(address,uint256,address)",

	// ======= AnyswapV5Router ===================
	"87cc6e2f": "anySwapFeeTo(address,uint256)",
	"825bb13c": "anySwapIn(bytes32,address,address,uint256,uint256)",
	"25121b76": "anySwapIn(bytes32[],address[],address[],uint256[],uint256[])",
	"0175b1c4": "anySwapInAuto(bytes32,address,address,uint256,uint256)",
	"52a397d5": "anySwapInExactTokensForNative(bytes32,uint256,uint256,address[],address,uint256,uint256)",
	"2fc1e728": "anySwapInExactTokensForTokens(bytes32,uint256,uint256,address[],address,uint256,uint256)",
	"3f88de89": "anySwapInUnderlying(bytes32,address,address,uint256,uint256)",
	"241dc2df": "anySwapOut(address,address,uint256,uint256)",
	"dcfb77b1": "anySwapOut(address[],address[],uint256[],uint256[])",
	"65782f56": "anySwapOutExactTokensForNative(uint256,uint256,address[],address,uint256,uint256)",
	"6a453972": "anySwapOutExactTokensForNativeUnderlying(uint256,uint256,address[],address,uint256,uint256)",
	"4d93bb94": "anySwapOutExactTokensForNativeUnderlyingWithPermit(address,uint256,uint256,address[],address,uint256,uint8,bytes32,bytes32,uint256)",
	"c8e174f6": "anySwapOutExactTokensForNativeUnderlyingWithTransferPermit(address,uint256,uint256,address[],address,uint256,uint8,bytes32,bytes32,uint256)",
	"0bb57203": "anySwapOutExactTokensForTokens(uint256,uint256,address[],address,uint256,uint256)",
	"d8b9f610": "anySwapOutExactTokensForTokensUnderlying(uint256,uint256,address[],address,uint256,uint256)",
	"99cd84b5": "anySwapOutExactTokensForTokensUnderlyingWithPermit(address,uint256,uint256,address[],address,uint256,uint8,bytes32,bytes32,uint256)",
	"9aa1ac61": "anySwapOutExactTokensForTokensUnderlyingWithTransferPermit(address,uint256,uint256,address[],address,uint256,uint8,bytes32,bytes32,uint256)",
	"a5e56571": "anySwapOutNative(address,address,uint256)",
	"edbdf5e2": "anySwapOutUnderlying(address,address,uint256,uint256)",
	"8d7d3eea": "anySwapOutUnderlyingWithPermit(address,address,address,uint256,uint256,uint8,bytes32,bytes32,uint256)",
	"1b91a934": "anySwapOutUnderlyingWithTransferPermit(address,address,address,uint256,uint256,uint8,bytes32,bytes32,uint256)",
	"99a2f2d7": "cID()",
	"5b7b018c": "changeMPC(address)",
	"456862aa": "changeVault(address,address)",
	"701bb891": "depositNative(address,address)",
	"c45a0155": "factory()",
	"85f8c259": "getAmountIn(uint256,uint256,uint256)",
	"054d50d4": "getAmountOut(uint256,uint256,uint256)",
	"1f00ca74": "getAmountsIn(uint256,address[])",
	"d06ca61f": "getAmountsOut(uint256,address[])",
	//"f75c2664": "mpc()",
	"ad615dec": "quote(uint256,uint256,uint256)",
	"8fd903f5": "wNATIVE()",
	"832e9492": "withdrawNative(address,uint256,address)",

	// ======= AnyCallProxy ===================
	"32f29022": "anyCall(address,address[],bytes[],address[],uint256[],uint256)",
	"5a11d475": "anyCall(address[],bytes[],address[],uint256[],uint256)",
	"b63b38d0": "applyMPC()",
	//"99a2f2d7": "cID()",
	//"5b7b018c": "changeMPC(address)",
	//"6a42b8f8": "delay()",
	"160f1053": "delayMPC()",
	"d6b0f484": "disableWhitelist()",
	"cdfb2b4e": "enableWhitelist()",
	"b510c235": "encode(string,bytes)",
	"ce82a36f": "encodePermit(address,address,uint256,uint256,uint8,bytes32,bytes32)",
	"99c1bc92": "encodeTransferFrom(address,address,uint256)",
	"09fd8212": "isInWhitelist(address)",
	//"f75c2664": "mpc()",
	"f830e7b4": "pendingMPC()",
	"f59c3708": "whitelist(address,bool)",
	"51fb012d": "whitelistEnabled()",
}

func parseEthTx(tx *types.Transaction) {
	to := tx.To()
	if to == nil {
		log.Println("the tx is creating contract.")
	}
	txData := tx.Data()
	if len(txData) < 4 {
		log.Println("the tx is not calling contract.")
	} else if method, exist := knownContractMethods[hex.EncodeToString(txData[:4])]; exist {
		log.Printf("the tx is calling method => %v", method)
	}

}
