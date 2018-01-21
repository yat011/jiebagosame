package main

import (
	"fmt"
	"jiebagosame/jieba"
	"time"
)
func main() {

	tokenizer, _ := jieba.NewTokeniezer()
	tokenizer.Hello()
	start := time.Now()
	for i := 0; i < 100000 ; i++ {
		tokenizer.CutDagNoHMM([]rune("我来到北京清华大学"))
		tokenizer.CutDagNoHMM([]rune("他来到了网易杭研大厦"))
	}
	elapsed := time.Since(start)
	fmt.Printf("took %s\n", elapsed)
	//fmt.Println("hihi")
}