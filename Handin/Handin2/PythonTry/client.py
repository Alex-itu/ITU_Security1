import ssl
import socket
context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
context.load_verify_locations("certificate.pem") 

client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

with context.wrap_socket(client_socket, server_hostname="Alex") as secure_socket:
    secure_socket.connect(('Alex', 443))
    print("SSL connection established")
    secure_socket.send(b"Hello, secure server!")
    data = secure_socket.recv(1024)
    print("Received:", data.decode())

client_socket.close()