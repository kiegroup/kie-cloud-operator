package shared

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/pavel-v-chernykh/keystore-go"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"math/big"
	"math/rand"
	"time"
)

// GenerateKeystore returns a Java Keystore with a self-signed certificate
func GenerateKeystore(commonName, alias string, password []byte) []byte {
	cert, derPK, err := genCert(commonName)
	if err != nil {
		log.Error("Error generating certificate. ", err)
	}

	var chain []keystore.Certificate
	keyStore := keystore.KeyStore{
		alias: &keystore.PrivateKeyEntry{
			Entry: keystore.Entry{
				CreationDate: time.Now(),
			},
			PrivKey: derPK,
			CertChain: append(chain, keystore.Certificate{
				Type:    "X509",
				Content: cert,
			}),
		},
	}

	var b bytes.Buffer
	err = keystore.Encode(&b, keyStore, password)
	if err != nil {
		log.Error("Error encrypting and signing keystore. ", err)
	}

	return b.Bytes()
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
