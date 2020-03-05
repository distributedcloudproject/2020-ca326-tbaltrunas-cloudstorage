KEY=$1
CERT=$2 

echo "Key: $KEY"
echo "Cert: $CERT"

openssl genrsa -out $KEY 2048
openssl req -new -x509 -sha256 -key $KEY -out $CERT -days 365
