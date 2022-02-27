package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/Valorith/EQRaidAssist/player"
)

var (
	Players   []*player.Player // Players detected within the raid dump file
	Rebooting bool             = false
	KEY       string
)

// Returns the current player cache
func GetActivePlayers() []*player.Player {
	return Players
}

func AddPlayer(p *player.Player) error {
	if IsCachedPlayer(p.Name) {
		return fmt.Errorf("player %s is already cached", p.Name)
	}
	Players = append(Players, p)
	return nil
}

func ClearPlayers() {
	Players = nil
	fmt.Println("Cached players cleared...")
}

// Check if the provided characterName is in the list of players
func IsCachedPlayer(characterName string) bool {
	for _, p := range Players {
		if p.Name == characterName {
			return true
		}
	}
	return false
}

func Encrypt(stringToEncrypt string, keyString string) (string, error) {

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	fmt.Printf("[Debug] key length (%s->%s): %d\n", keyString, key, len(key))
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("encrypt: NewCipher: %s", err)
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("Encrypt: NewGCM: %s", err)
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("Encrypt: ReadFull: %s", err)
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func Decrypt(encryptedString string, keyString string) (string, error) {

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Decrypt: NewCipher: %s", err)
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("Decrypt: NewGCM: %s", err)
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("Decrypt: aesGCM.Open: %s", err)
	}

	return string(plaintext), nil
}
