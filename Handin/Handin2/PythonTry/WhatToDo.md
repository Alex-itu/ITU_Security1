Steps:
1. Creating the private key
    - Using openssl: openssl genrsa -out key.pem 2048
    - This creates a private key with a protective passphrase

2. Creating the CSR (certificate signing request) using the created key
    - Using: openssl req -new -key key.pem -out signreq.csr

3. Signing the certificate with the key
    - Using: openssl x509 -req -days 365 -in signreq.csr -signkey key.pem -out certificate.pem
    - Since it being signed with the same key that created it, it is called a "self signed" certificate
    - optional: see details: openssl x509 -text -noout -in certificate.pem