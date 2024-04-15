package test

import "math/rand"

func GenerateRandomBytes(numOfBytesToGenerate int) []byte {
	generatedBytes := make([]byte, numOfBytesToGenerate)
	_, err := rand.Read(generatedBytes)
	if err != nil {
		panic(err)
	}
	return generatedBytes
}
