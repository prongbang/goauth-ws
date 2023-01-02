# goauth-ws

Authorization Websocket with JWT Example

### Run server

```shell
go run .
```

### Open browser

```shell
http://localhost:8080/
```

### Test with API

```
curl http://localhost:8080/device/publish?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

Output

```json
{"id":"1e4832e7-1ffa-4cf4-b9d9-0b8eff286c52","name":"Temp"}
```

### Test with MQTTX

```json
url: wss://broker.emqx.io:8084/mqtt
username: emqx_test
password: emqx_test
topic: device
payload: {"id":"1e4832e7-1ffa-4cf4-b9d9-0b8eff286c52","name":"Temp"}
```