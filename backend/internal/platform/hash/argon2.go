package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Hasher defines password hashing primitives.
type Hasher interface {
	Hash(password string) (string, error)
	Compare(encodedHash, password string) error
}

// Argon2Hasher hashes passwords using Argon2id.
type Argon2Hasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// NewArgon2Hasher constructs a Hasher with the given parameters.
func NewArgon2Hasher(memory, iterations, saltLength, keyLength uint32, parallelism uint8) (*Argon2Hasher, error) {
	if memory == 0 || iterations == 0 || saltLength == 0 || keyLength == 0 || parallelism == 0 {
		return nil, errors.New("argon2 parameters must be greater than zero")
	}

	return &Argon2Hasher{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  saltLength,
		keyLength:   keyLength,
	}, nil
}

// Hash encodes the given password using Argon2id and returns the encoded hash string.
func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", h.memory, h.iterations, h.parallelism, b64Salt, b64Hash)
	return encoded, nil
}

// Compare checks whether the provided password matches the encoded hash.
func (h *Argon2Hasher) Compare(encodedHash, password string) error {
	params, salt, expectedHash, err := decodeHash(encodedHash)
	if err != nil {
		return err
	}

	computed := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	if subtle.ConstantTimeCompare(computed, expectedHash) == 0 {
		return errors.New("password mismatch")
	}
	return nil
}

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
}

func decodeHash(encoded string) (argonParams, []byte, []byte, error) {
	var params argonParams

	var version int
	var memory, iterations, parallelism uint32
	var saltB64, hashB64 string

	_, err := fmt.Sscanf(encoded, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", &version, &memory, &iterations, &parallelism, &saltB64, &hashB64)
	if err != nil {
		return params, nil, nil, fmt.Errorf("parse hash: %w", err)
	}

	if version != 19 {
		return params, nil, nil, errors.New("unsupported argon2 version")
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return params, nil, nil, fmt.Errorf("decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return params, nil, nil, fmt.Errorf("decode hash: %w", err)
	}

	params = argonParams{
		memory:      memory,
		iterations:  iterations,
		parallelism: uint8(parallelism),
		keyLength:   uint32(len(hash)),
	}

	return params, salt, hash, nil
}
