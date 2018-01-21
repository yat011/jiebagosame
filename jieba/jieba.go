package jieba

import (

	"os"
	"log"
	"bufio"
	"strings"
	"strconv"
	"math"
)

type Tokenizer struct{
	freq  map[string] int
	total int64

}
type routeItem struct{
	index int
	loglikehood float64
}

func NewTokeniezer() (*Tokenizer, error){
	freqDict := make(map[string]int)
	var total int64 = 0

	file, err := os.Open("./jieba/dict.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splits := strings.Split(line," ")
		value, err := strconv.ParseInt(splits[1], 10,64 )
		if err != nil{
			return nil, err
		}
		word := splits[0]
		runes := []rune(word)
		freqDict[word] = int(value)
		for indx, _ := range runes{
			subword := string(runes[:indx+1])
			if _, ok := freqDict[subword]; ok == false{
				freqDict[subword] = 0
			}
		}
		total += value
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	tokenizer := Tokenizer{freq:freqDict,total:total}
	return &tokenizer, nil
}

//re_han_default = re.compile("([\u4E00-\u9FD5a-zA-Z0-9+#&\._%]+)", re.U)

func (tokenizer * Tokenizer) Hello(){

}

func (Tokenizer) Cut(sentence string, cut_all bool,HMM bool) {

}

func (tokenizer * Tokenizer) getDAG(sentence []rune) map[int][]int{
	dag := make(map[int][]int)
	for k, _ := range sentence{
		templist := make([]int,0)
		word := string(sentence[k:k+1])
		freq, ok := tokenizer.freq[word]
		i := k
		for  i < len(sentence) && ok == true{
			if freq != 0 {
				templist = append(templist, i)
			}
			i += 1
			if i+1 > len(sentence){
				break
			}
			word = string(sentence[k:i+1])
			freq, ok = tokenizer.freq[word]
		}
		if len(templist) == 0 {
			templist = append(templist, k)
		}
		dag[k] = templist
	}
	return dag
}

func (tokenizer * Tokenizer) CutDagNoHMM(sentence []rune) []string{
	dag := tokenizer.getDAG(sentence)
	route := tokenizer.calMaxLogLikehoodRoute(sentence, dag)
	x := 0
	n := len(sentence)
	buffer := ""
	output := make([]string,0)
	for x < n {
		y := route[x].index + 1
		word := string(sentence[x:y])
		if checkWordIfEnglish(word) && len(word) == 1{
			buffer+=word
			x = y
		}else{
			if len(buffer) > 0 {
				output = append(output, buffer)
				buffer = ""
			}
			output = append(output, word)
			x = y
		}
	}
	if len(buffer) > 0 {
		output = append(output, buffer)
	}
	return output
}

func (tokenizer * Tokenizer) calMaxLogLikehoodRoute(sentence [] rune, dag map[int][]int) []routeItem {
	n := len(sentence)
	route := make([]routeItem,n+1)
	route[n] = routeItem{index:0, loglikehood:0}
	logtotal := math.Log(float64(tokenizer.total))
	for idx:=n-1; idx >= 0; idx-- {
		var maxLoglike float64 = -1e9
		var maxItem *routeItem = nil
		for _, x := range dag[idx]{
			freq , ok := tokenizer.freq[string(sentence[idx:x+1])]
			if ok == false || freq == 0{
				freq = 1
			}
			loglike := math.Log(float64(freq))
			loglike -= logtotal
			loglike += route[x+1].loglikehood
			if loglike >= maxLoglike{
				maxLoglike = loglike
				maxItem = &routeItem{loglikehood:loglike,index:x}
			}
		}
		route[idx] = *maxItem
	}
	return route

}



func SplitByChinese(text string) [] string{
	blocks := make([] string,0, len(text))
	count := 0
	currentInRange := true
	currentBlock := ""
	for _, r := range text{
		if checkCharacterIfInRange(r){
			if currentInRange {
				currentBlock += string(r)
			}else{
				if len(currentBlock) > 0 {
					blocks = append(blocks, currentBlock)
					count += 1
				}
				currentBlock = string(r)
				currentInRange = true
			}
		}else{
			if currentInRange {
				if len(currentBlock) > 0 {
					blocks = append(blocks, currentBlock)
					count += 1
				}
				currentBlock = string(r)
				currentInRange = false
			}else{
				currentBlock += string(r)
			}
		}
	}
	if len(currentBlock) > 0{
		blocks = append(blocks, currentBlock)
	}
	return blocks
}

func checkCharacterIfInRange(r rune) bool{
	//re_han_default = re.compile("([\u4E00-\u9FD5a-zA-Z0-9+#&\._%]+)", re.U)
	if r >= '\u4E00' && r <= '\u9FD5' {
		return true
	}
	if r >= rune('a') && r <= rune('z'){
		return true
	}
	if r >= rune('A') && r <= rune('Z'){
		return true
	}
	if r >= rune('0') && r <= rune('9'){
		return true
	}
	if r == '#' || r ==  '+' || r == '&' || r == '.' || r == '_' || r == '%' {
		return true
	}
	return false

}

func checkIfEnglish(r rune) bool{
	//re_han_default = re.compile("([\u4E00-\u9FD5a-zA-Z0-9+#&\._%]+)", re.U)
	if r >= rune('a') && r <= rune('z'){
		return true
	}
	if r >= rune('A') && r <= rune('Z'){
		return true
	}
	if r >= rune('0') && r <= rune('9'){
		return true
	}
	return false
}
func checkWordIfEnglish(word string) bool{
	for _, v := range word{
		if checkIfEnglish(v) == false{
			return false
		}
	}
	return true


}



func IsChinese(text string) bool {
	for _, r := range []rune(text) {
		if r < '\u4E00' || r > '\u9FD5' {
			return false
		}
	}
	return true
}


