package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vmihailenco/msgpack/v5"
)

func AddressToBytes(address string) ([]byte, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAddress, address)
	}

	address = strings.TrimPrefix(address, "0x")
	return hex.DecodeString(address)
}

// ActionHash calculates the hash of an action for signing purposes
func ActionHash(action interface{}, vaultAddress string, nonce uint64) ([]byte, error) {
	data, err := msgpack.Marshal(action)
	if err != nil {
		return nil, fmt.Errorf("marshalling action: %w", err)
	}

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	data = append(data, nonceBytes...)

	if vaultAddress == "" {
		data = append(data, 0)
	} else {
		data = append(data, 1)
		addrBytes, err := AddressToBytes(vaultAddress)
		if err != nil {
			return nil, fmt.Errorf("processing vault address: %w", err)
		}
		data = append(data, addrBytes...)
	}

	hash := crypto.Keccak256(data)
	return hash, nil
}

// ConstructPhantomAgent constructs a phantom agent data structure
func ConstructPhantomAgent(hash []byte, isMainnet bool) map[string]interface{} {
	source := "a" // mainnet
	if !isMainnet {
		source = "b" // testnet
	}

	return map[string]interface{}{
		"source":       source,
		"connectionId": hexutil.Encode(hash),
	}
}

// HashMessage computes the Ethereum signed message hash
// This is equivalent to the eth_sign RPC method
func HashMessage(message []byte) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))

	return crypto.Keccak256(append([]byte(prefix), message...))
}

// VerifySignature verifies that a signature is valid for a given message and address
func VerifySignature(address string, message []byte, sig Signature) (bool, error) {
	if !common.IsHexAddress(address) {
		return false, fmt.Errorf("%w: %s", ErrInvalidAddress, address)
	}

	addrBytes, err := AddressToBytes(address)
	if err != nil {
		return false, err
	}

	r, err := hexutil.Decode(sig.R)
	if err != nil {
		return false, fmt.Errorf("invalid R value: %w", err)
	}

	s, err := hexutil.Decode(sig.S)
	if err != nil {
		return false, fmt.Errorf("invalid S value: %w", err)
	}

	hash := HashMessage(message)

	pubKey, err := crypto.Ecrecover(hash, append(append(r, s...), sig.V))
	if err != nil {
		return false, fmt.Errorf("recovering public key: %w", err)
	}

	ecdsaPubKey, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return false, fmt.Errorf("unmarshaling public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*ecdsaPubKey)

	return bytes.Equal(recoveredAddr.Bytes(), addrBytes), nil
}

func GenerateRandomCloid() (Cloid, error) {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}

	return Cloid(hex.EncodeToString(randBytes)), nil
}
