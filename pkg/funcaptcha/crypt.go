package funcaptcha

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
)

type encryptionData struct {
	Ct string `json:"ct"`
	Iv string `json:"iv"`
	S  string `json:"s"`
}

func Encrypt(data string, key string) string {
	encData, _ := aesEncrypt(data, key)

	encDataJson, err := json.Marshal(encData)
	if err != nil {
		panic(err)
	}

	return string(encDataJson)
}

func aesEncrypt(content string, password string) (*encryptionData, error) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	key, iv, err := defaultEvpKDF([]byte(password), salt)

	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	cipherBytes := pKCS5Padding([]byte(content), aes.BlockSize)
	mode.CryptBlocks(cipherBytes, cipherBytes)

	//TODO: remove redundant code
	md5Hash := md5.New()
	salted := ""
	var dx []byte

	for i := 0; i < 3; i++ {
		md5Hash.Write(dx)
		md5Hash.Write([]byte(password))
		md5Hash.Write(salt)

		dx = md5Hash.Sum(nil)
		md5Hash.Reset()

		salted += hex.EncodeToString(dx)
	}

	cipherText := base64.StdEncoding.EncodeToString(cipherBytes)
	encData := &encryptionData{
		Ct: cipherText,
		Iv: salted[64 : 64+32],
		S:  hex.EncodeToString(salt),
	}
	return encData, nil
}

// https://stackoverflow.com/questions/27677236/encryption-in-javascript-and-decryption-with-php/27678978#27678978
// https://github.com/brix/crypto-js/blob/8e6d15bf2e26d6ff0af5277df2604ca12b60a718/src/evpkdf.js#L55
func evpKDF(password []byte, salt []byte, keySize int, iterations int, hashAlgorithm string) ([]byte, error) {
	var block []byte
	var hasher hash.Hash
	derivedKeyBytes := make([]byte, 0)
	switch hashAlgorithm {
	case "md5":
		hasher = md5.New()
	default:
		return []byte{}, errors.New("not implement hasher algorithm")
	}
	for len(derivedKeyBytes) < keySize*4 {
		if len(block) > 0 {
			hasher.Write(block)
		}
		hasher.Write(password)
		hasher.Write(salt)
		block = hasher.Sum([]byte{})
		hasher.Reset()

		for i := 1; i < iterations; i++ {
			hasher.Write(block)
			block = hasher.Sum([]byte{})
			hasher.Reset()
		}
		derivedKeyBytes = append(derivedKeyBytes, block...)
	}
	return derivedKeyBytes[:keySize*4], nil
}

func defaultEvpKDF(password []byte, salt []byte) (key []byte, iv []byte, err error) {
	// https://github.com/brix/crypto-js/blob/8e6d15bf2e26d6ff0af5277df2604ca12b60a718/src/cipher-core.js#L775
	keySize := 256 / 32
	ivSize := 128 / 32
	derivedKeyBytes, err := evpKDF(password, salt, keySize+ivSize, 1, "md5")
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return derivedKeyBytes[:keySize*4], derivedKeyBytes[keySize*4:], nil
}
func pKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func Decrypt(data string, password string, fallbackPass string) string {
	decDataText, err := AesDecrypt(data, password, fallbackPass)

	if err != nil {
		println(err.Error())
	}

	return decDataText
}

func AesDecrypt(baseText string, password string, fallbackPass string) (string, error) {
	encBytes, err := base64.StdEncoding.DecodeString(baseText)
	if err != nil {
		return "", err
	}
	var encData encryptionData
	err = json.Unmarshal(encBytes, &encData)
	if err != nil {
		return "", err
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(encData.Ct)
	if err != nil {
		return "", err
	}
	salt, err := hex.DecodeString(encData.S)
	if err != nil {
		return "", err
	}
	dstBytes := make([]byte, len(cipherBytes))
decrypt:
	key, _, err := DefaultEvpKDF([]byte(password), salt)
	iv, _ := hex.DecodeString(encData.Iv)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dstBytes, cipherBytes)
	result := PKCS5UnPadding(dstBytes)
	if !json.Valid(result) {
		if password == fallbackPass {
			return "", errors.New("decryption can't get the correct result")
		} else {
			password = fallbackPass
			goto decrypt
		}
	}
	return string(result), nil
}

// https://stackoverflow.com/questions/27677236/encryption-in-javascript-and-decryption-with-php/27678978#27678978
// https://github.com/brix/crypto-js/blob/8e6d15bf2e26d6ff0af5277df2604ca12b60a718/src/evpkdf.js#L55
func EvpKDF(password []byte, salt []byte, keySize int, iterations int, hashAlgorithm string) ([]byte, error) {
	var block []byte
	var hasher hash.Hash
	derivedKeyBytes := make([]byte, 0)
	switch hashAlgorithm {
	case "md5":
		hasher = md5.New()
	default:
		return []byte{}, errors.New("not implement hasher algorithm")
	}
	for len(derivedKeyBytes) < keySize*4 {
		if len(block) > 0 {
			hasher.Write(block)
		}
		hasher.Write(password)
		hasher.Write(salt)
		block = hasher.Sum([]byte{})
		hasher.Reset()

		for i := 1; i < iterations; i++ {
			hasher.Write(block)
			block = hasher.Sum([]byte{})
			hasher.Reset()
		}
		derivedKeyBytes = append(derivedKeyBytes, block...)
	}
	return derivedKeyBytes[:keySize*4], nil
}

func DefaultEvpKDF(password []byte, salt []byte) (key []byte, iv []byte, err error) {
	// https://github.com/brix/crypto-js/blob/8e6d15bf2e26d6ff0af5277df2604ca12b60a718/src/cipher-core.js#L775
	keySize := 256 / 32
	ivSize := 128 / 32
	derivedKeyBytes, err := EvpKDF(password, salt, keySize+ivSize, 1, "md5")
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return derivedKeyBytes[:keySize*4], derivedKeyBytes[keySize*4:], nil
}

// https://stackoverflow.com/questions/41579325/golang-how-do-i-decrypt-with-des-cbc-and-pkcs7
func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}
