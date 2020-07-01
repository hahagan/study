import socket
from time import ctime
import pyarrow as pa

host = ''
port = 11000
ADDR = (host, port)
BUFSIZ = 12536800

tcpSocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
tcpSocket.bind(ADDR)
#set the max number of tcp connection
tcpSocket.listen(5)

while True:
    print('waiting for connection...')
    clientSocket, clientAddr = tcpSocket.accept()
    print('conneted form: %s' %clientAddr[0])

    while True:
        try:
            data = clientSocket.recv(BUFSIZ*2)
        except IOError as e:
            print(e)
            clientSocket.close()
            break
        if not data:
            break

        data1 = pa.deserialize(data)
        print("qweqwe")
    print(data1)
    clientSocket.close()

tcpSocket.close()