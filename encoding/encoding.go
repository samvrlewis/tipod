package encoding

import (
	"encoding/base32"
	"strings"
)

const (
	// Each base32 digit is 5 bits. We can have 63 digits per subdomain - so
	// 63*5 bits = 39.375 bytes
	SUBDOMAIN_B32_CHUNK_SIZE = 39
)

func ToDomainPrefix(data []byte) (string, error) {

	var domainPrefixBuilder strings.Builder

	for i := 0; i < len(data); i += SUBDOMAIN_B32_CHUNK_SIZE {
		end := i + SUBDOMAIN_B32_CHUNK_SIZE

		if end > len(data) {
			end = len(data)
		}

		b32Encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data[i:end])

		domainPrefixBuilder.WriteString(b32Encoded)
		domainPrefixBuilder.WriteString(".")
	}

	return strings.ToLower(domainPrefixBuilder.String()), nil
}

func FromDomainPrefix(domainPrefix string) ([]byte, error) {
	parts := strings.Split(strings.ToUpper(domainPrefix), ".")
	data := make([]byte, 0)

	for _, part := range parts {
		dataPart, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(part)

		if err != nil {
			return nil, err
		}

		data = append(data, dataPart...)
	}

	return data, nil
}

func ToTxtData(data []byte) (string, error) {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data), nil
}

func FromTxtData(data string) ([]byte, error) {
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(data)
}
