package jiebagosame

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

type matchFunc func(rune) bool

func NewTokeniezer(dictPath string) (*Tokenizer, error){
	freqDict := make(map[string]int)
	var total int64 = 0

	file, err := os.Open(dictPath)
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

func (tokenizer * Tokenizer) AddWord(word string, freq int, tag string){
	if freq <= 0 {
		freq = tokenizer.suggestFreq(word, false)
	}
	tokenizer.freq[word] = freq
	tokenizer.total += int64(freq)
	wordRune := []rune(word)
	for indx, _ := range wordRune{
		subword := string(wordRune[:indx+1])
		if _, ok := tokenizer.freq[subword]; ok == false{
			tokenizer.freq[subword] = 0
		}
	}
}

func (tokenizer *Tokenizer) suggestFreq(word string, tune bool) int{
	var freq float64 = 1
	var ok bool = false
	var wordFreq int
	wordFreq, ok = tokenizer.freq[word]
	if !ok{
		wordFreq = 1
	}
	for _, seg := range tokenizer.Cut(word,false, false){
		segFreq, ok :=  tokenizer.freq[seg]
		if !ok{
			segFreq = 1
		}
		freq *= float64(segFreq) / float64(tokenizer.total)
	}
	sugFreq := int(freq * float64(tokenizer.total))+1
	if wordFreq > sugFreq{
		return wordFreq
	}else{
		return sugFreq
	}
}

func (tokenizer * Tokenizer) Cut(sentence string, cut_all bool,HMM bool) []string {
	if cut_all{
		panic("not yet implemented")
	}
	if HMM {
		panic("HMM not yet implemented")
	}
	blocks := splitByHanDefault([]rune(sentence))
	result := make([] string ,0)
	for _, blk := range blocks{
		if len(blk) == 0{
			continue
		}
		runeBlk := []rune(blk)
		if checkWordIfHanDefeault(runeBlk){
			result = append(result, tokenizer.cutDagNoHMM(runeBlk)...)
		}else{
			for _, tempBlk := range splitByFunc(runeBlk, checkIfCharacterWhitespace)	{
				if checkIfWordWhiteSpace(tempBlk){
					result = append(result, string(tempBlk))
				}else{
					for _, sub := range tempBlk{
						result = append(result, string(sub))
					}
				}
			}

		}
	}
	return result
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

func (tokenizer * Tokenizer) cutDagNoHMM(sentence []rune) []string{
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

func splitByHanDefault(text []rune) []string{
	blocks := make([] string,0, len(text))
	count := 0
	currentInRange := true
	currentBlock := ""
	for _, r := range text{
		if checkCharacterIfInHanDefeault(r){
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

func checkWordIfHanDefeault(word []rune)bool{
	for _, w := range word{
		if checkCharacterIfInHanDefeault(w)==false{
			return false
		}
	}
	return true
}

func checkCharacterIfInHanDefeault(r rune) bool{
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

func checkIfCharacterWhitespace(r rune)bool{
	if r == rune('\r') || r == rune('\n') || r == rune('\t') || r == rune(' ') {
		return true
	}
	return false
}

func checkIfWordWhiteSpace(s string) bool{
	for _, r := range s{
		if !checkIfCharacterWhitespace(r){
			return false
		}
	}
	return true

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


func splitByFunc(text []rune, m matchFunc ) []string{
	blocks := make([] string,0, len(text))
	count := 0
	currentInRange := true
	currentBlock := ""
	for _, r := range text{
		if m(r){
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





