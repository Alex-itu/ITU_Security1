openssl req -new -newkey rsa:2048 -keyout ca.key -x509 -sha256 -days 365 -out ca.crt

openssl genrsa -out server.key 2048

openssl req -new -key server.key -out server.csr -config server.cnf

openssl req -noout -text -in server.csr

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 365 -sha256 -extfile server.cnf -extensions v3_ext