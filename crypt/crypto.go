package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
)

func New(passphare []byte, usersalt []byte) (key []byte, salt []byte, err error) {
	if len(passphare) < 1 {
		err = fmt.Errorf("need more than that for passphare")
		return
	}
	if usersalt == nil {
		salt = make([]byte, 8)
		if _, err := rand.Read(salt); err != nil {
			log.Fatalf("can't get random salt:%v", err)
		}
	} else {
		salt = usersalt
	}
	key = pbkdf2.Key(passphare, salt, 100, 32, sha256.New)
	return
}

func Encrypt(plaintext []byte, key []byte) (encrypted []byte, err error) {
	ivBytes := make([]byte, 12)
	if _, err := rand.Read(ivBytes); err != nil {
		log.Fatalf("can't initialize crypto:%v", err)
	}
	b, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return
	}
	encrypted = aesgcm.Seal(nil, ivBytes, plaintext, nil)
	encrypted = append(ivBytes, encrypted...)
	return
}

func Decrypt(encrypted []byte, key []byte) (plaintext []byte, err error) {
	if len(encrypted) < 13 {
		err = fmt.Errorf("incorrect passphare")
		return
	}
	b, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return
	}
	plaintext, err = aesgcm.Open(nil, encrypted[:12], encrypted[12:], nil)
	return
}

func NewArgon2(passphare []byte, usersalt []byte) (aead cipher.AEAD, salt []byte, err error) {
	if len(passphare) < 1 {
		err = fmt.Errorf("need more than that for passphare")
		return
	}

	if usersalt == nil {
		salt = make([]byte, 8)
		if _, err := rand.Read(salt); err != nil {
			log.Fatalf("can't get random salt:%v", err)
		}
	} else {
		salt = usersalt
	}
	aead, err = chacha20poly1305.NewX(argon2.IDKey(passphare, salt, 1, 64*1024, 4, 32))
	return
}

func EncryptChaCha(plaintext []byte, aead cipher.AEAD) (encrypted []byte, err error) {
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(plaintext)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}
	encrypted = aead.Seal(nonce, nonce, plaintext, nil)
	return
}

func DecryptChaCha(encryptedMsg []byte, aead cipher.AEAD) (encrypted []byte, err error) {
	if len(encryptedMsg) < aead.NonceSize() {
		err = fmt.Errorf("ciphertext too short")
		return
	}
	nonce, ciphertext := encryptedMsg[:aead.NonceSize()], encryptedMsg[aead.NonceSize():]
	encrypted, err = aead.Open(nil, nonce, ciphertext, nil)
	return
}
