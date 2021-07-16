package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"runtime"
	"strconv"
	"strings"
)

var (
	ErrWrongPass = errors.New("wrong password")
	ErrInvalidHash = errors.New("hash is invalid")
)

var argonParallelism = uint8(runtime.NumCPU())
const (
	argonMemory = uint32(64 * 1024)
	argonIterations = uint32(1)
	argonHashLength = uint32(16)
	argonSaltLength = 16
)

type Crypto interface {
	Rand(len int) ([]byte, error)

	hashPass(pass string) (string, error)
	verifyPass(pass string, hash string) error
}

type cryptoS struct{}

func NewCrypto() Crypto {
	return cryptoS{}
}

func (cryptoS) Rand(len int) ([]byte, error) {
	randBytes := make([]byte, len)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	return randBytes, nil
}

func (c cryptoS) hashPass(pass string) (string, error) {
	salt, err := c.Rand(argonSaltLength)
	if err != nil {
		return "", err
	}

	hashedPass := argon2.IDKey([]byte(pass), salt, argonIterations, argonMemory, argonParallelism, argonHashLength)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hashedPass),
	), nil
}

func (c cryptoS) verifyPass(pass string, hash string) error {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return ErrInvalidHash
	}

	params := strings.Split(parts[3], ",")
	memory, err := strconv.Atoi(params[0][2:])
	if err != nil {
		return err
	}

	iterations, err := strconv.Atoi(params[1][2:])
	if err != nil {
		return err
	}

	parallelism, err := strconv.Atoi(params[2][2:])
	if err != nil {
		return err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return err
	}

	actualHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return err
	}

	hashed := argon2.IDKey(
		[]byte(pass),
		salt,
		uint32(iterations),
		uint32(memory),
		uint8(parallelism),
		uint32(len(actualHash)),
	)

	if bytes.Equal(hashed, actualHash) {
		return nil
	} else {
		return ErrWrongPass
	}
}
