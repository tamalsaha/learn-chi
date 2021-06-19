# learn-chi

## Test Injector
```
go run main.go
```

http://localhost:3333/inject?name=tamal
http://localhost:3333/k8s

## TODOs

- [ ] Decide how to handle return values
- [ ] Handle Macaron style return values

- [ ] RequestID using https://github.com/oklog/ulid

## k8s apiserver

- k8s.io/apiserver/pkg/endpoints/installer.go
- k8s.io/apiserver/pkg/endpoints/groupversion.go
- k8s.io/apiserver/pkg/endpoints/handlers/create.go
---
- k8s.io/apiserver/pkg/endpoints/handlers/negotiation/negotiate.go
- k8s.io/apiserver/pkg/endpoints/handlers/responsewriters/status.go
- k8s.io/apiserver/pkg/endpoints/handlers/responsewriters/writers.go
