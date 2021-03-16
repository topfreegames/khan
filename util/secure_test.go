package util

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/wildlife-studios/crypto"
)

var encryptionKey []byte = []byte("a91j39s833hncy61alp0qb6e0s72pql14")
var data string = "some_data_test"

var _ = Describe("Security package", func() {
	Describe("EncryptData", func() {
		It("Should encrypt with XChacha and encode to base64", func() {
			xChacha := crypto.NewXChacha()
			encrypted, err := xChacha.Encrypt([]byte(data), encryptionKey[:32])
			Expect(err).NotTo(HaveOccurred())

			encryptedData, err := EncryptData(data, encryptionKey[:32])
			Expect(err).NotTo(HaveOccurred())

			encoded := base64.StdEncoding.EncodeToString(encrypted)
			Expect(encryptedData).To(Equal(encoded))

		})

		It("Should return in error case encryptionKey length is different than 32 bytes", func() {
			_, err := EncryptData(data, encryptionKey[:31])
			Expect(err).To(HaveOccurred())

			if _, ok := err.(*TokenSizeError); !ok {
				Fail("Error is not TokenSizeError")
			}

			Expect(err.Error()).To(Equal("The key length is different than 32"))

			_, err = EncryptData(data, encryptionKey[:33])
			Expect(err).To(HaveOccurred())

			if _, ok := err.(*TokenSizeError); !ok {
				Fail("Error is not TokenSizeError")
			}

			Expect(err.Error()).To(Equal("The key length is different than 32"))
		})
	})

	Describe("DecryptData", func() {
		It("Should decode with base64 after decrypt with XChacha", func() {
			encryptedData, err := EncryptData(data, encryptionKey)
			Expect(err).NotTo(HaveOccurred())

			cipheredData, err := base64.StdEncoding.DecodeString(encryptedData)
			Expect(err).NotTo(HaveOccurred())

			xChacha := crypto.NewXChacha()
			decrypted, err := xChacha.Decrypt([]byte(cipheredData), encryptionKey[:32])
			Expect(err).NotTo(HaveOccurred())

			decryptedData, err := DecryptData(encryptedData, encryptionKey[:32])
			Expect(err).NotTo(HaveOccurred())

			Expect(decryptedData).To(Equal(fmt.Sprintf("%s", decrypted)))
			Expect(decryptedData).To(Equal(data))

		})

		It("Should return in error case encryptionKey length is less than 32 bytes", func() {
			_, err := DecryptData(data, encryptionKey[:31])
			Expect(err).To(HaveOccurred())

			if _, ok := err.(*errors.TokenSizeError); !ok {
				Fail("Error is not TokenSizeError")
			}

			Expect(err.Error()).To(Equal("The key length is different than 32"))

			_, err = DecryptData(data, encryptionKey[:33])
			Expect(err).To(HaveOccurred())

			if _, ok := err.(*errors.TokenSizeError); !ok {
				Fail("Error is not TokenSizeError")
			}

			Expect(err.Error()).To(Equal("The key length is different than 32"))
		})
	})
})
