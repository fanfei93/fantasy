package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	path := "/Users/fanfei/fantasy/go-test/filepath-walk/a"
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		if path == "/Users/fanfei/fantasy/go-test/filepath-walk/a/b" {
			return errors.New("111")
		}
		return nil
	})
}
