# import http.server

# import ssl


# httpd = http.server.HTTPServer(('localhost', 443), http.server.SimpleHTTPRequestHandler)

# httpd.socket = ssl.wrap_socket (httpd.socket, certfile="./certificate.pem", server_side=True, ssl_version=ssl.PROTOCOL_TLS)

# httpd.serve_forever()

import ssl
import socket

# Create an SSL context for the server
context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
context.load_cert_chain(certfile="certificate.pem", keyfile="key.pem")

# Create a standard TCP socket
server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server_socket.bind(('Alex', 443)) 
server_socket.listen(3)

print("Server listening on port 443...")

while True:
    client_socket, addr = server_socket.accept()
    print(f"Connection from {addr}")

    # Wrap the socket with SSL
    with context.wrap_socket(client_socket, server_side=True) as secure_socket:
        print("SSL connection established")
        data = secure_socket.recv(1024)
        print("Received:", data.decode())
        secure_socket.send(b"Hello, secure client!")

    client_socket.close()