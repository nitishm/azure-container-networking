# Test Server streaming using gRPC

The sample program starts a server streaming and client gRPC service. The server runs as a go routine listening on port 8080 while the client loops forever receiving a stream of messages from the server after a connection has been established (initiated by the client by sending the client ID through an RPC call).

The client logs the recieved stream of messages to the console.

## Running the program

### Vendor the github.com/fatih/struct package

```terminal
go mod vendor
```

```terminal
❯ go run main.go
[Client ID: test-client] Received event type APPLY object type IPSET: 
IPSET-0: 172.17.0.0/2
IPSET-1: 172.17.0.1/30
IPSET-2: 172.17.0.2/14
IPSET-3: 172.17.0.3/22
IPSET-4: 172.17.0.4/2
IPSET-5: 172.17.0.5/12
IPSET-6: 172.17.0.6/18
IPSET-7: 172.17.0.7/24
IPSET-8: 172.17.0.8/16
IPSET-9: 172.17.0.9/8
[Client ID: test-client] Received event type APPLY object type IPSET: 
IPSET-0: 172.17.0.0/13
IPSET-1: 172.17.0.1/31
IPSET-2: 172.17.0.2/5
IPSET-3: 172.17.0.3/3
IPSET-4: 172.17.0.4/16
IPSET-5: 172.17.0.5/21
IPSET-6: 172.17.0.6/22
IPSET-7: 172.17.0.7/11
IPSET-8: 172.17.0.8/10
IPSET-9: 172.17.0.9/5
... ctrl-c to exit ...
^C2021/12/06 12:06:49 received signal: interrupt
[Client ID: test-client] Received event type APPLY object type IPSET: 
IPSET-0: 172.17.0.0/30
IPSET-1: 172.17.0.1/21
IPSET-2: 172.17.0.2/17
IPSET-3: 172.17.0.3/4
IPSET-4: 172.17.0.4/30
IPSET-5: 172.17.0.5/22
IPSET-6: 172.17.0.6/31
IPSET-7: 172.17.0.7/16
IPSET-8: 172.17.0.8/12
IPSET-9: 172.17.0.9/15
exit status 1
```
