<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/g/telemost"><img src=".github/images/godoc.svg"/></a>
  <a href="https://kaos.sh/c/telemost"><img src="https://kaos.sh/c/telemost.svg" alt="Coverage Status" /></a>
  <a href="https://kaos.sh/y/telemost"><img src="https://kaos.sh/y/df079c09ea4a4471aeabaebf960fa194.svg" alt="Codacy badge" /></a>
  <a href="https://kaos.sh/w/telemost/ci"><img src="https://kaos.sh/w/telemost/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/telemost/codeql"><img src="https://kaos.sh/w/telemost/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#usage-example">Usage example</a> • <a href="#ci-status">CI Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`telemost` is client for [Yandex.Telemost API](https://yandex.ru/dev/telemost/doc/ru/access).

### Usage example

```go
package main

import (
  "fmt"

  "github.com/essentialkaos/telemost"
)

func main() {
  api, err := telemost.NewClient("myToken1234")

  if err != nil {
    fmt.Println(err)
    return
  }

  // Create new conference
  cnf, err := api.Create(&telemost.Conference{})

  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Printf("Conference %s created, join — %s\n", cnf.ID, cnf.JoinURL)
}
```

### CI Status

| Branch | Status |
|--------|----------|
| `master` | [![CI](https://kaos.sh/w/telemost/ci.svg?branch=master)](https://kaos.sh/w/telemost/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/telemost/ci.svg?branch=develop)](https://kaos.sh/w/telemost/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/.github/blob/master/CONTRIBUTING.md).

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://kaos.dev"><img src="https://raw.githubusercontent.com/essentialkaos/.github/refs/heads/master/images/ekgh.svg"/></a></p>
