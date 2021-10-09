# Redmap

[![Go Reference](https://pkg.go.dev/badge/github.com/livingsilver94/redmap.svg)](https://pkg.go.dev/github.com/livingsilver94/redmap) [![Go Report Card](https://goreportcard.com/badge/livingsilver94/redmap)](https://goreportcard.com/report/livingsilver94/redmap)

Redmap is a general purpose Go module to convert structs to map of strings and vice versa, when possible.

One particular use case of Redmap is serializing and deserializing Redis object, which stores everything as strings.

### Goals

- Keep the API similar to `encoding/json`: we do all hate learning yet another library
- API stability
- Excellent test coverage

## Installation

In a project using Go modules, run this in your terminal application:

```bash
go get github.com/livingsilver94/redmap
```

When the download finishes, you'll find Redmap under the unsurprising name of `redmap`.

## Usage

[This example](https://play.golang.org/p/sIcwTP2zAzJ) shows the simplest usage of Redmap. Usage of struct tags and unmarshaling pitfalls are better illustrated in the documentation.

```go
package main

import (
	"fmt"
	"github.com/livingsilver94/redmap"
)

type MyStruct struct {
	AString string
	AnInt   int
}

func main() {
	// Struct to map.
	myS := MyStruct{AString: "universe", AnInt: 42}
	mp, err := redmap.Marshal(myS)
	fmt.Println(mp, err)

	// Map to struct.
	var mySIsBack MyStruct
	err = redmap.Unmarshal(mp, &mySIsBack)
	fmt.Println(mySIsBack, err)
}
```
