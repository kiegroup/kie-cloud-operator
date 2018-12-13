package shared

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"testing"

	keystore "github.com/pavel-v-chernykh/keystore-go"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestResourceRequirements(t *testing.T) {
	reqs := GetResourceRequirements(map[string]map[v1.ResourceName]string{"Limits": {v1.ResourceMemory: "1Gi", v1.ResourceCPU: "2"}, "Requests": {v1.ResourceMemory: "500Mi"}})
	log.V(1).Info(fmt.Sprintf("Resource Requirements: %v", reqs))
	assert.Equal(t, *reqs.Limits.Memory(), resource.MustParse("1Gi"))
	assert.Equal(t, *reqs.Limits.Cpu(), resource.MustParse("2"))
	assert.Equal(t, *reqs.Requests.Memory(), resource.MustParse("500Mi"))
}

func TestGenerateKeystore(t *testing.T) {
	alias := "test"
	password := GeneratePassword(8)
	assert.EqualValues(t, 8, len(password))
	defer Zeroing(password)

	commonName := "test-https"
	keyBytes := GenerateKeystore(commonName, alias, password)
	keyStore, err := keystore.Decode(bytes.NewReader(keyBytes), password)
	assert.Nil(t, err)

	derKey := keyStore[alias].(*keystore.PrivateKeyEntry).PrivKey
	_, err = x509.ParsePKCS8PrivateKey(derKey)
	assert.Nil(t, err)

	cert := keyStore[alias].(*keystore.PrivateKeyEntry).CertChain[0].Content
	certificate, err := x509.ParseCertificate(cert)
	assert.Nil(t, err)
	assert.Equal(t, commonName, certificate.Subject.CommonName)
}
