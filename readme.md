##Goer-Shortlink 
a grpc url shortener server
-
- ###Features
1. accepts long url and returns a short link.
2. fetch url from database by looking at the shortid.

NB:- see /proto/urlshortener.proto file to see the service and entities.


### Run
- ``cd server && go build .``
supported flags are tls, port, certfile, keyfile

### Stack:
- go-grpc
- postgres
- zap logger

### TODO
 still collating and making design decisions.