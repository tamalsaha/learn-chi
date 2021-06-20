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

## Status / Error

- Gitea: https://github.com/go-gitea/gitea/blob/v1.14.3/modules/context/api.go#L32-L122

- form validation
  - https://github.com/go-playground/form
  - https://github.com/go-playground/validator/blob/master/_examples/simple/main.go

- k8s
 - Status Error type: https://github.com/kubernetes/apimachinery/blob/v0.21.1/pkg/apis/meta/v1/types.go#L648-L898
 - Error creators: https://github.com/kubernetes/apimachinery/blob/v0.21.1/pkg/api/errors/errors.go
 - error to APIStatus: https://github.com/kubernetes/apiserver/blob/release-1.21/pkg/endpoints/handlers/responsewriters/status.go
 - Write Error: https://github.com/kubernetes/apiserver/blob/release-1.21/pkg/endpoints/handlers/responsewriters/writers.go#L278-L296

## k8s apiserver

- k8s.io/apiserver/pkg/endpoints/installer.go
- k8s.io/apiserver/pkg/endpoints/groupversion.go
- k8s.io/apiserver/pkg/endpoints/handlers/create.go
---
- k8s.io/apiserver/pkg/endpoints/handlers/negotiation/negotiate.go
- k8s.io/apiserver/pkg/endpoints/handlers/responsewriters/status.go
- k8s.io/apiserver/pkg/endpoints/handlers/responsewriters/writers.go
---
**Convert using Hub**
- sigs.k8s.io/controller-runtime/pkg/webhook/conversion/conversion.go
