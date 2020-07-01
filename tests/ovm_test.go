package tests

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

var chainConfig params.ChainConfig

func init() {
	chainConfig = params.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      new(big.Int),
		ByzantiumBlock:      new(big.Int),
		ConstantinopleBlock: new(big.Int),
		DAOForkBlock:        new(big.Int),
		DAOForkSupport:      false,
		EIP150Block:         new(big.Int),
		EIP155Block:         new(big.Int),
		EIP158Block:         new(big.Int),
	}
}

const GAS_LIMIT = 15000000

var ZERO_ADDRESS = common.HexToAddress("0000000000000000000000000000000000000000")

func TestSetExecutionContext(t *testing.T) {
	initCode, _ := hex.DecodeString("6080604052348015600f57600080fd5b5060b28061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032")
	state := newState()

	codeBefore := state.GetCode(common.HexToAddress("65486c8ec9167565ebd93c94ed04f0f71d1b5137"))

	// Now apply a simple message with contract deployment

	returnValue, gasUsed, failed, err := applyMessageToState(*state, ZERO_ADDRESS, GAS_LIMIT, initCode)

	// Verify the transition occurred
	code := state.GetCode(common.HexToAddress("65486c8ec9167565ebd93c94ed04f0f71d1b5137"))

	fmt.Println("Hello!")
	fmt.Println("Code before:", codeBefore, "Code After:", code)

	fmt.Println("Return val:", returnValue, "Gas used:", gasUsed, "Failed:", failed, "Error:", err)

	// expectedDeployedBytecode := common.FromHex("6080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032")
	// if !bytes.Equal(code[:], expectedDeployedBytecode) {
	// 	t.Errorf("Expected %020x; got %020x", expectedDeployedBytecode, code[:])
	// }
}

func TestSloadAndStore(t *testing.T) {
	rawStateManagerAbi, _ := ioutil.ReadFile("./StateManagerABI.json")
	stateManagerAbi, _ := abi.JSON(strings.NewReader(string(rawStateManagerAbi)))
	state := newState()

	address := common.HexToAddress("9999999999999999999999999999999999999999")
	key := [32]byte{}
	value := [32]byte{}
	copy(key[:], []byte("hello"))
	copy(value[:], []byte("world"))

	storeCalldata, _ := stateManagerAbi.Pack("setStorage", address, key, value)
	getCalldata, _ := stateManagerAbi.Pack("getStorage", address, key)

	call(t, state, vm.StateManagerAddress, storeCalldata)
	getStorageReturnValue, _ := call(t, state, vm.StateManagerAddress, getCalldata)

	if !bytes.Equal(value[:], getStorageReturnValue) {
		t.Errorf("Expected %020x; got %020x", value[:], getStorageReturnValue)
	}
}

func TestCreate(t *testing.T) {
	initCode, _ := hex.DecodeString("6080604052348015600f57600080fd5b5060b28061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032")
	rawStateManagerAbi, _ := ioutil.ReadFile("./StateManagerABI.json")
	stateManagerAbi, _ := abi.JSON(strings.NewReader(string(rawStateManagerAbi)))
	state := newState()

	address := common.HexToAddress("9999999999999999999999999999999999999999")
	callerAddress := common.HexToAddress("42")

	deployContractCalldata, _ := stateManagerAbi.Pack("deployContract", address, initCode, true, callerAddress)
	createdContractAddr, _ := call(t, state, vm.StateManagerAddress, deployContractCalldata)
	expectedDeployedBytecode := common.FromHex("6080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032")
	deployedBytecode := state.GetCode(common.BytesToAddress(createdContractAddr[12:]))
	if !bytes.Equal(deployedBytecode, expectedDeployedBytecode) {
		t.Errorf("Expected %020x; got %020x", expectedDeployedBytecode, deployedBytecode)
	}
}

func TestGetAndIncrementNonce(t *testing.T) {
	rawStateManagerAbi, _ := ioutil.ReadFile("./StateManagerABI.json")
	stateManagerAbi, _ := abi.JSON(strings.NewReader(string(rawStateManagerAbi)))
	state := newState()

	address := common.HexToAddress("9999999999999999999999999999999999999999")

	getNonceCalldata, _ := stateManagerAbi.Pack("getOvmContractNonce", address)
	incrementNonceCalldata, _ := stateManagerAbi.Pack("incrementOvmContractNonce", address)

	getStorageReturnValue1, _ := call(t, state, vm.StateManagerAddress, getNonceCalldata)

	expectedReturnValue1 := makeUint256WithUint64(0)
	if !bytes.Equal(getStorageReturnValue1, expectedReturnValue1) {
		t.Errorf("Expected %020x; got %020x", expectedReturnValue1, getStorageReturnValue1)
	}

	call(t, state, vm.StateManagerAddress, incrementNonceCalldata)
	getStorageReturnValue2, _ := call(t, state, vm.StateManagerAddress, getNonceCalldata)

	expectedReturnValue2 := makeUint256WithUint64(1)
	if !bytes.Equal(getStorageReturnValue2, expectedReturnValue2) {
		t.Errorf("Expected %020x; got %020x", expectedReturnValue2, getStorageReturnValue2)
	}
}

func TestGetCodeContractAddress(t *testing.T) {
	rawStateManagerAbi, _ := ioutil.ReadFile("./StateManagerABI.json")
	stateManagerAbi, _ := abi.JSON(strings.NewReader(string(rawStateManagerAbi)))
	state := newState()

	address := common.HexToAddress("9999999999999999999999999999999999999999")

	getCodeContractAddressCalldata, _ := stateManagerAbi.Pack("getCodeContractAddress", address)

	getCodeContractAddressReturnValue, _ := call(t, state, vm.StateManagerAddress, getCodeContractAddressCalldata)

	if !bytes.Equal(getCodeContractAddressReturnValue[12:], address.Bytes()) {
		t.Errorf("Expected %020x; got %020x", getCodeContractAddressReturnValue[12:], address.Bytes())
	}
}

func TestAssociateCodeContract(t *testing.T) {
	rawStateManagerAbi, _ := ioutil.ReadFile("./StateManagerABI.json")
	stateManagerAbi, _ := abi.JSON(strings.NewReader(string(rawStateManagerAbi)))
	state := newState()

	address := common.HexToAddress("9999999999999999999999999999999999999999")

	getCodeContractAddressCalldata, _ := stateManagerAbi.Pack("associateCodeContract", address, address)

	_, err := call(t, state, vm.StateManagerAddress, getCodeContractAddressCalldata)
	if err != nil {
		t.Errorf("Failed to call associateCodeContract: %s", err)
	}
}

func TestGetCodeContractBytecode(t *testing.T) {
	state := newState()
	initCode, _ := hex.DecodeString("6080604052348015600f57600080fd5b5060b28061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032")
	rawStateManagerAbi, _ := ioutil.ReadFile("./StateManagerABI.json")
	stateManagerAbi, _ := abi.JSON(strings.NewReader(string(rawStateManagerAbi)))
	address := common.HexToAddress("9999999999999999999999999999999999999999")
	callerAddress := common.HexToAddress("42")
	deployContractCalldata, _ := stateManagerAbi.Pack("deployContract", address, initCode, true, callerAddress)
	call(t, state, vm.StateManagerAddress, deployContractCalldata)
	getCodeContractBytecodeCalldata, _ := stateManagerAbi.Pack("getCodeContractBytecode", address)
	getCodeContractBytecodeReturnValue, _ := call(t, state, vm.StateManagerAddress, getCodeContractBytecodeCalldata)
	deployedByteCode := common.FromHex("0000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000b26080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032000000000000000000000000000000000000")
	if !bytes.Equal(getCodeContractBytecodeReturnValue, deployedByteCode) {
		t.Errorf("Expected %020x; got %020x", deployedByteCode, getCodeContractBytecodeReturnValue)
	}
}

func TestGetCodeContractHash(t *testing.T) {
	state := newState()
	initCode, _ := hex.DecodeString("6080604052348015600f57600080fd5b5060b28061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032")
	rawStateManagerAbi, _ := ioutil.ReadFile("./StateManagerABI.json")
	stateManagerAbi, _ := abi.JSON(strings.NewReader(string(rawStateManagerAbi)))
	address := common.HexToAddress("9999999999999999999999999999999999999999")
	callerAddress := common.HexToAddress("42")
	deployContractCalldata, _ := stateManagerAbi.Pack("deployContract", address, initCode, true, callerAddress)
	call(t, state, vm.StateManagerAddress, deployContractCalldata)
	getCodeContractBytecodeCalldata, _ := stateManagerAbi.Pack("getCodeContractHash", address)
	getCodeContractBytecodeReturnValue, _ := call(t, state, vm.StateManagerAddress, getCodeContractBytecodeCalldata)
	expectedCreatedCodeHash := crypto.Keccak256(common.FromHex("6080604052348015600f57600080fd5b506004361060285760003560e01c80639b0b0fda14602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8060008084815260200190815260200160002081905550505056fea265627a7a7231582053ac32a8b70d1cf87fb4ebf5a538ea9d9e773351e6c8afbc4bf6a6c273187f4a64736f6c63430005110032"))
	if !bytes.Equal(getCodeContractBytecodeReturnValue, expectedCreatedCodeHash) {
		t.Errorf("Expected %020x; got %020x", getCodeContractBytecodeReturnValue, expectedCreatedCodeHash)
	}
}

func makeUint256WithUint64(num uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, num)
	val := append(make([]byte, 24), b[:]...)
	return val
}

func newState() *state.StateDB {
	db := state.NewDatabase(rawdb.NewMemoryDatabase())
	state, _ := state.New(common.Hash{}, db)
	return state
}

func applyMessageToState(state state.StateDB, to common.Address, gasLimit uint64, data []byte) ([]byte, uint64, bool, error) {
	header := &types.Header{
		Number:     big.NewInt(0),
		Difficulty: big.NewInt(0),
	}
	gasPool := core.GasPool(100000000)
	// Default from address
	from := common.HexToAddress("8888888888888888888888888888888888888888")
	// Generate the message
	message := types.Message{}
	if to == ZERO_ADDRESS {
		// Check if to the ZERO_ADDRESS, if so, make it nil
		message = types.NewMessage(
			from,
			nil,
			0,
			big.NewInt(0),
			gasLimit,
			big.NewInt(0),
			data,
			false,
		)
	} else {
		// Otherwise we actually use the `to` field!
		message = types.NewMessage(
			from,
			&to,
			0,
			big.NewInt(0),
			gasLimit,
			big.NewInt(0),
			data,
			false,
		)
	}

	context := core.NewEVMContext(message, header, nil, &from)
	evm := vm.NewEVM(context, &state, &chainConfig, vm.Config{})

	returnValue, gasUsed, failed, err := core.ApplyMessage(evm, message, &gasPool)

	state.Finalise(true)

	return returnValue, gasUsed, failed, err
}

func call(t *testing.T, state *state.StateDB, address common.Address, callData []byte) ([]byte, error) {
	returnValue, _, err := runtime.Call(address, callData, &runtime.Config{
		State:       state,
		ChainConfig: &chainConfig,
	})

	return returnValue, err
}
