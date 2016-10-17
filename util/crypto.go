package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

var block cipher.Block

func SetAES256Cipher(buf string) error {

	key := []byte(buf)

	var err error
	block, err = aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("Create Cipher:%s", err)
	}
	return nil
}

func EncryptAES256(t string) (string, error) {
	if block == nil {
		return "", fmt.Errorf("Call SetAES256Cipher()")
	}
	plain := []byte(t)
	cipher := make([]byte, len(plain))
	block.Encrypt(cipher, plain)
	return string(cipher), nil
}

func DecryptAES256(t string) (string, error) {
	if block == nil {
		return "", fmt.Errorf("Call SetAES256Cipher()")
	}
	cipher := []byte(t)
	text := make([]byte, len(cipher))
	block.Decrypt(text, cipher)
	return string(text), nil
}
