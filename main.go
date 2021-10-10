package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FindLine struct {
	LineNum int
	LintStr string
}

type FindFile struct {
	Name      string
	FindLines []FindLine
}

type FindFiles struct {
	files []*FindFile
}

var mem *FindFiles
var once sync.Once

func GetFileList(path string) ([]string, error) {
	return filepath.Glob(path)
}

func FindText(path, word string, ch chan *FindFile) {
	findFile := &FindFile{Name: path}
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("파일을 찾을 수 없습니다. err: %s\n", err)
		ch <- nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		s := scanner.Text()
		if strings.Contains(s, word) {
			findFile.FindLines = append(findFile.FindLines, FindLine{LineNum: lineNum, LintStr: s})
		}
		lineNum++
	}
	ch <- findFile
}

func (f *FindFiles) FindTextFromFiles(path, word string, ch chan *FindFile) {
	fileList, err := GetFileList(path)
	fileCnt := len(fileList)
	recvCnt := 0
	if err != nil {
		fmt.Printf("파일을 찾을 수 없습니다. err: %s\n", err)
		return
	}

	for _, file := range fileList {
		go FindText(file, word, ch)
	}
	for file := range ch {
		f.files = append(f.files, file)
		recvCnt++
		if fileCnt == recvCnt {
			break
		}
	}
}

func (f *FindFiles) PrintResult() {
	for _, v := range f.files {
		fmt.Println(v.Name)
		fmt.Println("----------------------------------")
		for _, line := range v.FindLines {
			fmt.Printf("%d\t%s\n", line.LineNum, line.LintStr)
		}
		fmt.Println("----------------------------------")
		fmt.Println()
	}
}

func InitMem() {
	once.Do(func() {
		mem = &FindFiles{}
	})
}

func main() {
	args := os.Args
	fileCh := make(chan *FindFile)
	if len(args) < 3 {
		fmt.Println("찾을 단어나 파일을 입력해주세요.")
		return
	}
	InitMem()
	word, filepaths := args[1], args[2:]
	for _, path := range filepaths {
		mem.FindTextFromFiles(path, word, fileCh)
	}
	mem.PrintResult()
}
