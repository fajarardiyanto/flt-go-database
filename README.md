### Go Module Logger
Database modules, this modules designed to concurrent safe

Faltar Database use :
- [ElasticSearch v7.17.1](github.com/elastic/go-elasticsearch) as Elasticsearch.

### Installation
```sh
go get github.com/fajarardiyanto/flt-go-database
```

###### Upgrading to the latest version
```sh
go get -u github.com/fajarardiyanto/flt-go-database
```

###### Upgrade or downgrade with tag version if available
```sh
go get -u github.com/fajarardiyanto/flt-go-database@v1.0.0
```

### Usage
```go
package main

import (
	"github.com/fajarardiyanto/flt-go-database/lib"
)

func main() {
	db := lib.NewLib()
	db.Init("Test Database Modules")
}

```

#### Run Example
```sh
make help
```

#### Tips
Maybe it would be better to do some basic code scanning before pushing to the repository.
```sh
# for *.nix users just run gosec.sh
# curl is required
# more information https://github.com/securego/gosec
make scan
```