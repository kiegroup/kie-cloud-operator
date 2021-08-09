package shared

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"

	"github.com/pavel-v-chernykh/keystore-go/v4"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GenerateKeystore returns a Java Keystore with a self-signed certificate
func GenerateKeystore(commonName string, password []byte) ([]byte, error) {
	var b bytes.Buffer
	certificate, derPK, err := genCert(commonName)
	if err != nil {
		return []byte{}, err
	}
	keyStore := keystore.New(keystore.WithOrderedAliases())
	pkeIn := keystore.PrivateKeyEntry{
		CreationTime: time.Now(),
		PrivateKey:   derPK,
		CertificateChain: []keystore.Certificate{
			{
				Type:    "X509",
				Content: certificate,
			},
		},
	}
	if err := keyStore.SetPrivateKeyEntry(constants.KeystoreAlias, pkeIn, password); err != nil {
		return []byte{}, err
	}
	if err := keyStore.Store(&b, password); err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}

func IsValidKeyStoreSecret(secret corev1.Secret, keystoreCN string, keyStorePassword []byte) (bool, error) {
	if secret.Data[constants.KeystoreName] != nil {
		return IsValidKeyStore(keystoreCN, keyStorePassword, secret.Data[constants.KeystoreName])
	}
	return false, nil
}

func IsValidKeyStore(keystoreCN string, keyStorePassword, keyStoreData []byte) (bool, error) {
	keyStore := keystore.New(keystore.WithOrderedAliases())
	// FIX err == nil or something else!
	if err := keyStore.Load(bytes.NewReader(keyStoreData), keyStorePassword); err != nil {
		return false, err
	}
	if ok := keyStore.IsPrivateKeyEntry(constants.KeystoreAlias); !ok {
		return false, nil
	}
	pke, err := keyStore.GetPrivateKeyEntry(constants.KeystoreAlias, keyStorePassword)
	if err != nil {
		return false, err
	}
	return commonNameExists(keystoreCN, pke.CertificateChain)
}

func commonNameExists(keystoreCN string, certChain []keystore.Certificate) (bool, error) {
	for _, certEntry := range certChain {
		cert, err := x509.ParseCertificate(certEntry.Content)
		if err != nil {
			return false, err
		}
		if cert.Subject.CommonName == keystoreCN {
			return true, nil
		}
	}
	return false, nil
}

// GenerateTruststore returns a Java Truststore with a Trusted CA bundle
func GenerateTruststore(caBundle []byte) ([]byte, error) {
	var b bytes.Buffer
	trustStore, err := createTruststoreObject(caBundle)
	if err != nil {
		return []byte{}, err
	}
	if err := trustStore.Store(&b, []byte(constants.TruststorePwd)); err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}

func createTruststoreObject(caBundle []byte) (keystore.KeyStore, error) {
	trustStore := keystore.New(keystore.WithOrderedAliases())
	if ok, err := appendCertsFromPEM(caBundle, &trustStore); !ok {
		if err != nil {
			return keystore.KeyStore{}, err
		}
	}
	return trustStore, nil
}

func appendCertsFromPEM(pemCerts []byte, s *keystore.KeyStore) (ok bool, err error) {
	for i := 0; i < len(pemCerts); i++ {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		if err := s.SetTrustedCertificateEntry(cert.Issuer.ToRDNSequence().String(), keystore.TrustedCertificateEntry{
			CreationTime: time.Now(),
			Certificate: keystore.Certificate{
				Type:    "X509",
				Content: cert.Raw,
			},
		}); err != nil {
			return false, err
		}
	}
	return true, nil
}

func IsValidTruststoreSecret(secret corev1.Secret, caBundle []byte) (bool, error) {
	if secret.Data[constants.TruststoreName] != nil {
		return IsValidTruststore(caBundle, secret.Data[constants.TruststoreName])
	}
	return false, nil
}

func IsValidTruststore(caBundle, keyStoreData []byte) (bool, error) {
	existingtTrustStore := keystore.New(keystore.WithOrderedAliases())
	if err := existingtTrustStore.Load(bytes.NewReader(keyStoreData), []byte(constants.TruststorePwd)); err != nil {
		return false, err
	}
	trustStore, err := createTruststoreObject(caBundle)
	if err != nil {
		return false, err
	}
	existingtTrustAliases := existingtTrustStore.Aliases()
	trustAliases := trustStore.Aliases()
	if len(trustAliases) != len(existingtTrustAliases) {
		return false, nil
	}
	for _, alias := range existingtTrustAliases {
		existingCertEntry, err := existingtTrustStore.GetTrustedCertificateEntry(alias)
		if err != nil {
			return false, err
		}
		trustCertEntry, err := trustStore.GetTrustedCertificateEntry(alias)
		if err != nil {
			return false, err
		}
		if !reflect.DeepEqual(existingCertEntry.Certificate, trustCertEntry.Certificate) {
			return false, nil
		}
	}
	return true, nil
}

// ????????????????
// any way to use openshift's CA for signing instead ??
func genCert(commonName string) (cert []byte, derPK []byte, err error) {
	sAndI := pkix.Name{
		CommonName: commonName,
		//OrganizationalUnit: []string{"Engineering"},
		//Organization:       []string{"RedHat"},
		//Locality:           []string{"Raleigh"},
		//Province:           []string{"NC"},
		//Country:            []string{"US"},
	}

	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Error("Error getting serial number. ", err)
		return nil, nil, err
	}

	ca := &x509.Certificate{
		Subject:            sAndI,
		Issuer:             sAndI,
		SignatureAlgorithm: x509.SHA256WithRSA,
		PublicKeyAlgorithm: x509.ECDSA,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(10, 0, 0),
		SerialNumber:       serialNumber,
		SubjectKeyId:       sha256.New().Sum(nil),
		IsCA:               true,
		// BasicConstraintsValid: true,
		// ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		// KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		log.Error("create key failed. ", err)
		return nil, nil, err
	}

	cert, err = x509.CreateCertificate(crand.Reader, ca, ca, &priv.PublicKey, priv)
	if err != nil {
		log.Error("create cert failed. ", err)
		return nil, nil, err
	}

	derPK, err = x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Error("Marshal to PKCS8 key failed. ", err)
		return nil, nil, err
	}

	return cert, derPK, nil
}

// GeneratePassword returns an alphanumeric password of the length provided
func GeneratePassword(length int) []byte {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	for i := 1; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}

	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})

	return buf
}

// GetEnvVar returns the position of the EnvVar found by name
func GetEnvVar(envName string, env []corev1.EnvVar) int {
	for pos, v := range env {
		if v.Name == envName {
			return pos
		}
	}
	return -1
}

func EnvVarSet(env corev1.EnvVar, envList []corev1.EnvVar) bool {
	for _, e := range envList {
		if env.Name == e.Name && env.Value == e.Value {
			return true
		}
	}
	return false
}

// EnvOverride replaces or appends the provided EnvVar to the collection
func EnvOverride(dst, src []corev1.EnvVar) []corev1.EnvVar {
	for _, cre := range src {
		pos := GetEnvVar(cre.Name, dst)
		if pos != -1 {
			dst[pos] = cre
		} else {
			dst = append(dst, cre)
		}
	}
	return dst
}

// EnvVarCheck checks whether the src and dst []EnvVar have the same values
func EnvVarCheck(dst, src []corev1.EnvVar) bool {
	for _, denv := range dst {
		if !EnvVarSet(denv, src) {
			return false
		}
	}
	for _, senv := range src {
		if !EnvVarSet(senv, dst) {
			return false
		}
	}
	return true
}

func GetNamespacedName(object metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
	}
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// ValidateRouteHostname validates the hostname provided by the user
// see: https://github.com/openshift/router/blob/release-4.6/pkg/router/controller/unique_host.go#L231
func ValidateRouteHostname(r string) field.ErrorList {
	specPath := field.NewPath("spec")
	hostPath := specPath.Child("host")
	result := field.ErrorList{}
	if len(r) < 1 {
		log.Debugf("%s is empty, no custom hostname will be configured", hostPath)
		return result
	}

	if len(kvalidation.IsDNS1123Subdomain(r)) != 0 {
		result = append(result, field.Invalid(hostPath, r, "host must conform to DNS 952 subdomain conventions"))
	}
	segments := strings.Split(r, ".")
	for _, s := range segments {
		errs := kvalidation.IsDNS1123Label(s)
		for _, e := range errs {
			result = append(result, field.Invalid(hostPath, r, e))
		}
	}
	return result
}
