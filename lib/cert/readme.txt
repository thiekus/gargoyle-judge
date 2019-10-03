You can place pair of server key here. By default, gymaster will use server.crt and server.key
if was set to "useTLS" to true.

If you don't have yet, you can build self signed key pair (only for local uses):

openssl genrsa -out server.key 2048
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
