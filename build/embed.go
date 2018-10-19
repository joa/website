package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const path = "./static"

func main() {
	fs, err := ioutil.ReadDir(path)

	if err != nil {
		panic(err)
	}

	out, err := os.Create("bindata.go")

	if err != nil {
		panic(err)
	}

	if _, err := io.WriteString(out, "package main \n\nconst (\n"); err != nil {
		panic(err)
	}

	for _, f := range fs {
		if _, err := out.Write([]byte(f.Name()[0:strings.LastIndex(f.Name(), ".")] + " = `")); err != nil {
			panic(err)
		}

		f, _ := os.Open(filepath.Join(path, f.Name()))

		if _, err := io.Copy(out, f); err != nil {
			panic(err)
		}

		if _, err := io.WriteString(out, "`\n"); err != nil {
			panic(err)
		}
	}

	if _, err := io.WriteString(out, ")\n"); err != nil {
		panic(err)
	}
}
