To run this program you must have openSSL, and use the following commands to set up a certificate:

Create RSA key
openssl genrsa -out server.key 2048

Create Certificate from RSA key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365 -addext "subjectAltName = DNS:localhost"


Running program
-- server (start with this)
go run ./hos/hospital.go

-- clients (make 3 clients)
go run ./clients/client.go Portnumber

like this: 
client 1: go run ./clients/client.go 8551
client 2: go run ./clients/client.go 8552
client 3: go run ./clients/client.go 8553