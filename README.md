# adaptd
[![Go Report Card](https://goreportcard.com/badge/github.com/dadamssolutions/adaptd)](https://goreportcard.com/report/github.com/dadamssolutions/adaptd) [![GoDoc](https://godoc.org/github.com/dadamssolutions/adaptd?status.svg)](https://godoc.org/github.com/dadamssolutions/adaptd)

Adapters to add middleware to HTTP Handlers.

### Installing

Use `go get`:

```
go get github.com/dadamssolutions/adaptd
```

Or, in `go.mod`:

```
require (
    github.com/dadamssolutions/adaptd
)
```

## Examples

```go
import (
    "net/http"
    "github.com/dadamssolutions/adaptd"
)

func main() {
    // Index handler should enure that HTTPS is used
    http.Handle("/", adaptd.EnsureHTTPS(false)(indexHandler))
    // Login handler should use HTTPS and handle GET and POST requests
    loginHandler = adaptd.Apapt(loginHandler,
                    adaptd.EnsureHTTPS(false),
                    adaptd.GetAndOtherRequest(loginPostHandler, http.MethodPost))
    http.Handle("/login", loginHandler)

    http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil)
}
```

## Contributing

Submit a pull request.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags).

## Authors

* **Donnie Adams, Owner, dadams solutions llc** - *Initial work* - [dadams solutions llc](https://github.com/dadamssolutions)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

* [Mat Ryer](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)

