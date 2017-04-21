package main

import (
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type xmlFileInfo struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	MD5     string   `xml:"md5,attr"`
}

type xmlDirInfo struct {
	XMLName xml.Name `xml:"folder"`

	Files []*xmlFileInfo `xml:"file"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	pathPtr := flag.String("path", "", "the path to make md5.")
	outputPtr := flag.String("output", "md5.xml", "the filename to output. xml format.")
	flag.Parse()

	absPath := getAbsPath(*pathPtr)
	outputFilename := getAbsPath(*outputPtr)

	fmt.Println("path to make: ", absPath)
	fmt.Println("output filename:", outputFilename)

	filesList := list.New()

	if err := filepath.Walk(absPath, func(_path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			fmt.Println("Dir:", _path)
		} else {
			rel, err := filepath.Rel(absPath, _path)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("File:", rel)
			filesList.PushBack(rel)
		}

		return nil
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(filesList)

	var dirsInfo xmlDirInfo

	dirsInfo.Files = make([]*xmlFileInfo, filesList.Len())
	i := 0
	for it := filesList.Front(); nil != it; it = it.Next() {
		m, err := md5file(filepath.Join(absPath, it.Value.(string)))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		dirsInfo.Files[i] = &xmlFileInfo{Name: it.Value.(string), MD5: m}
		i++
	}

	//d := &xmlDirInfo{Files: []*xmlFileInfo{{Name: "a.txt", MD5: "dddd"}}}
	output, err := xml.MarshalIndent(dirsInfo, "", "	")
	if err != nil {
		fmt.Println(err)
		return
	}
	output = append([]byte(xml.Header), output...)
	//fmt.Println(string(output))

	ioutil.WriteFile(outputFilename, output, os.ModeAppend)

	fmt.Println("Done.")
}

func getAbsPath(_path string) string {

	if _path == "" {
		_path = "."
	}

	if !filepath.IsAbs(_path) {
		curDir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		return filepath.Join(curDir, _path)
	}
	return _path
}

func md5file(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	m := md5.New()
	io.Copy(m, f)
	f.Close()
	return hex.EncodeToString(m.Sum(nil)), nil
}
