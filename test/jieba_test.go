package test

import (
	"fmt"
	"time"
	"github.com/yat011/jiebagosame"
	"testing"
	"os"
	"bufio"
	"strings"
)

func TestJieba(t * testing.T) {

	tokenizer, err := jiebagosame.NewTokeniezer("../dict.txt")
	if err != nil{
		panic(err)
	}
	file, _ :=os.Open("./example.txt")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	content := make([]string,0)
	for scanner.Scan(){
		line := scanner.Text()
		content = append(content, line)
	}


	result := make([][]string, 0)
	start := time.Now()
	for _ ,line := range content{
		result = append(result,tokenizer.Cut(line, false, false))
	}
	elapsed := time.Since(start)
	fmt.Printf("took %s\n", elapsed)
	//fmt.Println("hihi")

	ansFile, aErr := os.Create("./go_ans.txt")
	defer ansFile.Close()
	if aErr != nil{
		panic(aErr)
	}
	for _, ans := range result{
		fmt.Fprintf(ansFile,"%s\r\n", strings.Join(ans, "|"))
	}
}