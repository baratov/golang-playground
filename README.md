# golang-playground

##Usage:

###Via client 
    examples in `client_test.go` file

###Via http

Headers: 
    Content-Type: application/json
    Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ= _(username:password)_
Path: 
    http://localhost:8080/api/v1/keys/{key}
Methods: 
    GET, POST, PUT, DELETE
Payload for POST and PUT:
    {"value":"some_value","ttl":3600000000000}    