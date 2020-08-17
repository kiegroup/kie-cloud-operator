package shared

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/pavel-v-chernykh/keystore-go/v4"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestEnvOverride(t *testing.T) {
	src := []corev1.EnvVar{
		{
			Name:  "test1",
			Value: "value1",
		},
		{
			Name:  "test2",
			Value: "value2",
		},
	}
	dst := []corev1.EnvVar{
		{
			Name:  "test1",
			Value: "valueX",
		},
		{
			Name:  "test3",
			Value: "value3",
		},
	}
	result := EnvOverride(src, dst)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, result[0], dst[0])
	assert.Equal(t, result[1], src[1])
	assert.Equal(t, result[2], dst[1])
}

func TestGetEnvVar(t *testing.T) {
	vars := []corev1.EnvVar{
		{
			Name:  "test1",
			Value: "value1",
		},
		{
			Name:  "test2",
			Value: "value2",
		},
	}
	pos := GetEnvVar("test1", vars)
	assert.Equal(t, 0, pos)

	pos = GetEnvVar("other", vars)
	assert.Equal(t, -1, pos)
}

func TestGenerateKeystore(t *testing.T) {
	password := GeneratePassword(8)
	assert.Len(t, password, 8)

	commonName := "test-https"
	keyBytes, err := GenerateKeystore(commonName, password)
	assert.Nil(t, err)
	ok, err := IsValidKeyStore(commonName, password, keyBytes)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestGenerateTruststore(t *testing.T) {
	caBundle, err := ioutil.ReadFile("test-" + constants.CaBundleKey)
	assert.Nil(t, err)
	assert.NotEmpty(t, caBundle)

	trust1, err := createTruststoreObject(caBundle)
	assert.Nil(t, err)
	certChainLen := 129
	assert.Len(t, trust1.Aliases(), certChainLen)

	trustBytes, err := GenerateTruststore(caBundle)
	assert.Nil(t, err)

	existingtTrustStore := keystore.New(keystore.WithOrderedAliases())
	err = existingtTrustStore.Load(bytes.NewReader(trustBytes), []byte(constants.TruststorePwd))
	assert.Nil(t, err)
	assert.Len(t, existingtTrustStore.Aliases(), certChainLen)

	ok, err := IsValidTruststore(caBundle, trustBytes)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestTruststoreInvalid(t *testing.T) {
	caBundle, err := ioutil.ReadFile("test-" + constants.CaBundleKey)
	assert.Nil(t, err)
	assert.NotEmpty(t, caBundle)

	trust1, err := createTruststoreObject(caBundle)
	assert.Nil(t, err)
	certChainLen := 129
	assert.Len(t, trust1.Aliases(), certChainLen)

	emptyCa := []byte{}
	trustBytes1, err := GenerateTruststore(caBundle)
	assert.Nil(t, err)
	ok, err := IsValidTruststore(emptyCa, trustBytes1)
	assert.False(t, ok)
	assert.Nil(t, err)

	trustBytes2, err := GenerateTruststore(emptyCa)
	assert.Nil(t, err)
	ok, err = IsValidTruststore(caBundle, trustBytes2)
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestEnvVarCheck(t *testing.T) {
	empty := []corev1.EnvVar{}
	a := []corev1.EnvVar{
		{Name: "A", Value: "1"},
	}
	b := []corev1.EnvVar{
		{Name: "A", Value: "2"},
	}
	c := []corev1.EnvVar{
		{Name: "A", Value: "1"},
		{Name: "B", Value: "1"},
	}

	assert.True(t, EnvVarCheck(empty, empty))
	assert.True(t, EnvVarCheck(a, a))

	assert.False(t, EnvVarCheck(empty, a))
	assert.False(t, EnvVarCheck(a, empty))

	assert.False(t, EnvVarCheck(a, b))
	assert.False(t, EnvVarCheck(b, a))

	assert.False(t, EnvVarCheck(a, c))
	assert.False(t, EnvVarCheck(c, b))
}
