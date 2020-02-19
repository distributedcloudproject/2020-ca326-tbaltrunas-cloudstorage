package comm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestComm(t *testing.T) {
	key1, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := listen(0)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		err := acceptListener(listener, key1)
		if err != nil {
			t.Fatal(err)
		}
	}()

	key2, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	cl, err := NewClientDial(listener.Addr().String(), key2)
	go cl.HandleConnection()
	if err != nil {
		t.Fatal(err)
	}
	m, err := cl.SendMessage("ping", "ping")
	if err != nil {
		t.Fatal(err)
	}
	if m[0].(string) != "pong" {
		t.Fatal("got", m, "wanted pong")
	}
}

func TestError(t *testing.T) {
	key1, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := listen(0)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		err := acceptListener(listener, key1)
		if err != nil {
			t.Fatal(err)
		}
	}()

	key2, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	cl, err := NewClientDial(listener.Addr().String(), key2)
	go cl.HandleConnection()
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		Input string

		Output1 string
		Output2 string
		Error   string
	}{
		{
			Input:   "test1:test2",
			Output1: "test1",
			Output2: "test2",
			Error:   "",
		},
		{
			Input:   "var:voom",
			Output1: "var",
			Output2: "voom",
			Error:   "",
		},
		{
			Input:   "test1:test2:test3",
			Output1: "",
			Output2: "",
			Error:   "invalid string",
		},
		{
			Input:   "test1test2",
			Output1: "",
			Output2: "",
			Error:   "invalid string",
		},
	}
	for _, testCase := range testCases {
		m, err := cl.SendMessage("split", testCase.Input)
		if err == nil && testCase.Error != "" {
			t.Errorf("case(%v).Error got nil; want %v", testCase.Input, testCase.Error)
		}
		if err != nil && err.Error() != testCase.Error {
			t.Errorf("case(%v).Error got %v; want %v", testCase.Input, err.Error(), testCase.Error)
		}
		if m[0].(string) != testCase.Output1 {
			t.Errorf("case(%v).Output1 got %v; want %v", testCase.Input, m[0].(string), testCase.Output1)
		}
		if m[1].(string) != testCase.Output2 {
			t.Errorf("case(%v).Output2 got %v; want %v", testCase.Input, m[1].(string), testCase.Output2)
		}
	}
}

func SplitByColonTwice(msg string) (part1 string, part2 string, err error) {
	s := strings.Split(msg, ":")
	if len(s) != 2 {
		return "", "", errors.New("invalid string")
	}
	return s[0], s[1], nil
}

func Testt(msg string) string {
	if msg == "ping" {
		return "pong"
	}
	return ""
}

func listen(port int) (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func acceptListener(listener net.Listener, key *rsa.PrivateKey) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		client, err := NewServerClient(conn, key)
		client.RegisterRequest("ping", Testt)
		client.RegisterRequest("split", SplitByColonTwice)
		if err != nil {
			return err
		}
		go func() {
			client.HandleConnection()
		}()
	}
}

func generateCert() ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("%v: generating key P256", err))
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 365)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Wolf Cola"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("%v: creating certificate", err))
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("%v: marshaling private key", err))
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}), nil
}

func generateKey() (*rsa.PrivateKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pri, nil
}
