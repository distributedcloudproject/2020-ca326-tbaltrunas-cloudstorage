package comm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
)

// NewClient creates a new client with an existing network connection.
func NewClient(conn net.Conn, key *rsa.PrivateKey) (Client, error) {
	client, err := newClient(conn, key)
	if err != nil {
		return nil, err
	}

	clientRandom := make([]byte, 8)
	_, err = rand.Read(clientRandom)
	if err != nil {
		return nil, err
	}
	err = sendClientRandom(conn, clientRandom)
	if err != nil {
		return nil, err
	}

	// Read encrypted part of server's key.
	encryptedBytes, err := readConn(conn)
	if err != nil {
		return nil, err
	}
	b, err := rsa.DecryptPKCS1v15(rand.Reader, client.privateKey, encryptedBytes)
	if err != nil {
		return nil, err
	}
	masterKey := append(b, clientRandom...)
	client.masterKey = masterKey
	err = client.createCypher()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewServerClient creates a new server client with an existing network connection.
// Server Client is the client that accepted the connection, instead of the one that initiated it.
// Server Client will generate the master key.
func NewServerClient(conn net.Conn, key *rsa.PrivateKey) (Client, error) {
	client, err := newClient(conn, key)
	if err != nil {
		return nil, err
	}

	clientRandom, err := readClientRandom(conn)
	if len(clientRandom) != 8 {
		return nil, errors.New(fmt.Sprintf("client random length: %d; expected 8", len(clientRandom)))
	}

	// Generate remaining 24 bytes of the key.
	b := make([]byte, 24)
	n, err := rand.Reader.Read(b)
	if err != nil {
		return nil, err
	}
	if n != 24 {
		return nil, errors.New("failed reading 24 bytes from input stream")
	}
	masterKey := append(b, clientRandom...)

	client.masterKey = masterKey
	// Send our part of the key, encrypted with the client's public key.
	encryptedB, err := rsa.EncryptPKCS1v15(rand.Reader, client.publicKey, b)
	if err != nil {
		return nil, err
	}
	err = writeConn(conn, encryptedB)
	if err != nil {
		return nil, err
	}
	err = client.createCypher()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func newClient(conn net.Conn, key *rsa.PrivateKey) (*client, error) {
	client := &client{}

	client.conn = conn
	client.messages = make(map[uint32]*message)
	client.requests = make(map[string]interface{})
	client.privateKey = key

	err := sendPublicKey(conn, &key.PublicKey)
	if err != nil {
		return nil, err
	}
	pub, err := readPublicKey(conn)
	if err != nil {
		return nil, err
	}
	client.publicKey = pub

	return client, nil
}

func (c *client) createCypher() error {
	ci, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(ci)
	if err != nil {
		return err
	}
	c.cipher = gcm
	return nil
}

func writeConn(conn net.Conn, data []byte) error {
	lengthBuffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBuffer, uint32(len(data)))

	_, err := conn.Write(lengthBuffer)
	if err != nil {
		return err
	}
	sentBytes := 0
	for sentBytes < len(data) {
		n, err := conn.Write(data[sentBytes:])
		if err != nil {
			return err
		}
		sentBytes += n
	}
	return nil
}

func readConn(conn net.Conn) ([]byte, error) {
	b := make([]byte, 4)
	n, err := conn.Read(b)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, errors.New("communication: read " + strconv.Itoa(n) + " bytes; wanted 4")
	}

	msgLength := binary.LittleEndian.Uint32(b)
	b = make([]byte, int(msgLength))

	var read uint32
	for read < msgLength {
		n, err = conn.Read(b[read:])
		if err != nil {
			return nil, err
		}
		read += uint32(n)
	}
	return b, nil
}

func publicKeyToID(key *rsa.PublicKey) (string) {
	pub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		fmt.Println(err)
		return "ERR"
	}
	sha := sha256.Sum256(pub)
	return hex.EncodeToString(sha[:])
}

func sendPublicKey(conn net.Conn, key *rsa.PublicKey) error {
	pub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}
	pubKey := []byte(hex.EncodeToString(pub))

	return writeConn(conn, pubKey)
}

func readPublicKey(conn net.Conn) (*rsa.PublicKey, error) {
	b, err := readConn(conn)
	if err != nil {
		return nil, err
	}

	bytes, err := hex.DecodeString(string(b))
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKIXPublicKey(bytes)
	if err != nil {
		return nil, err
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("communication: invalid public key")
	}

	return pub, nil
}

func sendClientRandom(conn net.Conn, rand []byte) error {
	if len(rand) != 8 {
		return errors.New("failed reading 4 bytes from input stream")
	}

	err := writeConn(conn, rand)
	if err != nil {
		return err
	}
	return nil
}

func readClientRandom(conn net.Conn) ([]byte, error) {
	return readConn(conn)
}