package main

import (
	"log"
	"os"
	"path/filepath"
)

var path = "/tmp/test/test2"

func init() {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			log.Fatalln("创建文件夹失败：", err.Error())
		}
	}
}

func main() {
	method2("1.txt")
}

func method1(name string) {
	filename := filepath.Join(path, name)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("打开文件失败:", err.Error())
	}
	file.Close()
}

func method2(name string) {
	filename := filepath.Join(path, name)
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalln("打开文件失败:", err.Error())
	}
	f.Close()
	fo, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		log.Fatalln("打开文件失败")
	}
	log.Println("打开文件成功")
	fo.Close()
}
