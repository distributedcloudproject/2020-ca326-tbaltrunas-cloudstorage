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

func TestCommTimeout(t *testing.T) {
	key1, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := listen(0)
	if err != nil {
		t.Fatal(err)
	}
	var clientServ Client
	go func() {
		for {
			var err error
			conn, err := listener.Accept()
			t.Logf("Accepted conn: %v.", conn)
			if err != nil {
				t.Fatal(err)
			}

			clientServ, err = NewServerClient(conn, key1)
			t.Logf("%v.", clientServ)
			clientServ.RegisterRequest("ping", Testt)
			if err != nil {
				t.Fatal(err)
			}
			go func() {
				clientServ.HandleConnection()
			}()
		}

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

	_, err = cl.SendMessage("ping", "pingtimeout")
    if err == nil || err.Error() != "Timeout" {
        t.Fatalf("Wanted a Timeout error. Got: %v.", err)
    }
}

func Testt(msg string) string {
	if msg == "ping" {
		return "pong"
	} else if msg == "pingtimeout" {
		for {
			fmt.Println("Timing out...")
			time.Sleep(10 * time.Second)
		}
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
		NotAfter: notAfter,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("%v: creating certificate", err))
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil,  errors.New(fmt.Sprintf("%v: marshaling private key", err))
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