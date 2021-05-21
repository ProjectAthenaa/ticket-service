package deob

import (
	"encoding/hex"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	enabled = false

	bit32 = int(math.Pow(float64(2), float64(32)))
	bit31 = int(math.Pow(float64(2), float64(31)))
	Hex = "0123456789abcdef"
	staticmatch = regexp.MustCompile("static")
	switchregex = regexp.MustCompile("(?:var )?(\\w{3}(?:\\[\\d+])?)=(\\d+);for\\(;.*?!==(.*?);\\)\\{switch\\((.*?)\\)")
	basicvardef1 = regexp.MustCompile("(\\w{4}\\.\\w{3})=\"?(\\w+)\"?;")
	basicvardef2 = regexp.MustCompile("(\\w{4}\\[\\d+])=\"?(\\w+)\"?;")
	basicvarmatch1 = regexp.MustCompile("(\\w{4}\\.\\w{3})[^(]")
	basicvarmatch2 = regexp.MustCompile("(\\w{4}\\[\\d+])[^(]")

	pushregex = regexp.MustCompile("(\\w{3}\\[\\d+])\\[\\w{3}\\[\\d+]]\\((.*?)\\);\\w{3}\\[\\d+]\\[\\w{3}\\[\\d+]]\\((.*?)\\);\\w{3}\\[\\d+]\\[\\w{3}\\[\\d+]]\\((.*?)\\);\\w{3}\\[\\d+]\\[\\w{3}\\[\\d+]]\\((.*?)\\);")
	webglchallengematch = regexp.MustCompile("(\\w{3})\\[\\d+],(\\w{3}\\[\\d+])")
	interestmatch = regexp.MustCompile("\\w{3}\\[\\d+]")
	webglfloats = regexp.MustCompile("float\\((.*?)\\)")
	aceg = []int{0, 0, 0, 16, 16, 16, 16, 16, 32, 32, 32, 32, 32, 48, 48, 48, 48, 48, 64, 64, 64, 64, 64, 80, 80, 80, 80, 80, 96, 96, 96, 96, 96, 112, 112, 112, 112, 112, 128, 128, 128, 128, 128, 144, 144, 144, 144, 144, 160, 160, 160, 160, 160, 176, 176, 176, 176, 176, 192, 192, 192, 192, 192, 208, 208, 208, 208, 208, 224, 224, 224, 224, 224, 240, 240, 240, 240, 240, 256, 256}
	bdfh = []int{0, 0, 0, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 6, 6, 6, 6, 6, 7, 7, 7, 7, 7, 8, 8, 8, 8, 8, 9, 9, 9, 9, 9, 10, 10, 10, 10, 10, 11, 11, 11, 11, 11, 12, 12, 12, 12, 12, 13, 13, 13, 13, 13, 14, 14, 14, 14, 14, 15, 15, 15, 15, 15, 16, 16}

	keymatch = regexp.MustCompile("(\\w{3}\\[\\d+])=\\[]")
	retvar = regexp.MustCompile("\\w{3}\\[\\w+]\\[\\w{3}\\[\\w+]]\\((\\w{3}\\[\\w+]),\\w{3}\\[\\w+]\\);")
	lastkeyregex = regexp.MustCompile("(\\w{3}\\[\\d+])=new")

	binarycaseregex = regexp.MustCompile("\\?(\\d+):(\\d+);break;")
	linearcaseregex = regexp.MustCompile("(\\d+);break;")

	replacementregexp 		= regexp.MustCompile("[\\s]+")
	digitmatch              = regexp.MustCompile("\\d+")
	keydigitmatch			= regexp.MustCompile("(-)? ?\\(?\\+?\"?(\\d+)")
	multipleindexmatcher    = regexp.MustCompile("(?:\\[\\d+])+")
	boolArrayInitialization = regexp.MustCompile("(\\d+),(\\d+),\\[(\\d+),(\\d+)]")
	boolarrmatch = regexp.MustCompile("\\w{3,4}\\.\\w{3}\\(\\)(?:\\[\\d+])+")

	stringarrregex = regexp.MustCompile("\\('(.{6})'")
	returnR = regexp.MustCompile("return \"(.*?)\"")
	splitR = regexp.MustCompile("\\('(.)'")
	shiftR = regexp.MustCompile("\\((-?\\d+),(-?\\d+)\\)")

	A5f = regexp.MustCompile("^(function [0-9a-zA-Z_$]+\\([0-9a-zA-Z_$]+,\\s*[0-9a-zA-Z_$]+\\)\\s*\\{)\\s*(\"use strict\";)([\\s\\S]*)$")

	obffuncregex = regexp.MustCompile("\\((function \\w{3}\\(\\w{3}\\).*?)\\)\\(\"(.*?)\"\\)")
	count = 0
)

type switchObj struct{
	Variable string
	Initialization int
	Escapes []int
	Cases map[int]caseObj
	Startindex int
	Endindex int
	Basescript string
}
type caseObj struct{
	Value int
	Body string
	Routeone int
	Routetwo int //omit if binary
	Type string //linear, binary
}
func conditionalwalker(matchcase, startcase int, escapes []int, cases map[int]caseObj) int {
	if startcase == matchcase{
		return 1
	}else if intsliceincludes(escapes, startcase){
		return 0
	}
	switch cases[startcase].Type {
	case "linear":
		return conditionalwalker(matchcase, cases[startcase].Routeone, escapes, cases)
	case "binary":
		escapes = append(escapes, startcase)
		return conditionalwalker(matchcase, cases[startcase].Routeone, escapes, cases) + conditionalwalker(matchcase, cases[startcase].Routetwo, escapes, cases)
	}
	return 0
}
func switchwalker(nextcase int, escapes []int, cases map[int]caseObj) string {
	//retstring := ""
	var retstring strings.Builder
	returns := 1
	for returns > 0{
		if intsliceincludes(escapes, nextcase){
			if _, ok := cases[nextcase]; !ok{
				return retstring.String()
			}else if cases[nextcase].Type == "binary"{
				return retstring.String()
			}
		}

		switch cases[nextcase].Type{
		case "binary":
			conditiontype := []int{conditionalwalker(nextcase, cases[nextcase].Routeone, escapes, cases), conditionalwalker(nextcase, cases[nextcase].Routetwo, escapes, cases)}
			if (conditiontype[0] == 0 && conditiontype[1] == 0) {
				retstring.WriteString(cases[nextcase].Body + fmt.Sprintf("{%s}{%s}", switchwalker(cases[nextcase].Routeone, escapes, cases), switchwalker(cases[nextcase].Routetwo, escapes, cases)))
				returns--
			} else if (conditiontype[0] > 0 && conditiontype[1] == 0) {
				escapes = append(escapes, nextcase)
				retstring.WriteString(cases[nextcase].Body + fmt.Sprintf("{%s}", switchwalker(cases[nextcase].Routeone, escapes, cases)))
				nextcase = cases[nextcase].Routetwo
			} else if (conditiontype[0] == 0 && conditiontype[1] > 0) {
				escapes = append(escapes, nextcase)
				retstring.WriteString(cases[nextcase].Body + fmt.Sprintf("{%s}", switchwalker(cases[nextcase].Routetwo, escapes, cases)))
				nextcase = cases[nextcase].Routeone
			} else if (conditiontype[0] > 0 && conditiontype[1] > 0) {
				escapes = append(escapes, nextcase)
				retstring.WriteString(cases[nextcase].Body + fmt.Sprintf("{{%s}{%s}}", switchwalker(cases[nextcase].Routeone, escapes, cases), switchwalker(cases[nextcase].Routetwo, escapes, cases)))
				returns--
			}
		case "linear":
			retstring.WriteString(cases[nextcase].Body)
			if intsliceincludes(escapes, nextcase){
				returns--
			}else{
				nextcase = cases[nextcase].Routeone
			}
		default:
			retstring.WriteString(cases[nextcase].Body)
			return retstring.String()
		}
	}
	return retstring.String()
}

func caseCreator(objin switchObj) switchObj {
	for _, caserstring := range strings.Split(objin.Basescript, "case ")[1:]{
		appendcase := caseObj{}
		casesplit := strings.SplitN(caserstring, ":", 2)
		caseindex, _ := strconv.Atoi(casesplit[0])
		appendcase.Value = caseindex
		nextroutes := strings.Split(casesplit[1], fmt.Sprintf("%s=", objin.Variable))
		switch len(nextroutes) {
		case 2:
			appendcase.Body = nextroutes[0]
			binarycases := binarycaseregex.FindStringSubmatch(nextroutes[1])
			switch(len(binarycases)){
			case 3:
				routeone, _ := strconv.Atoi(binarycases[1])
				routetwo, _ := strconv.Atoi(binarycases[2])
				appendcase.Routeone = routeone
				appendcase.Routetwo = routetwo
				appendcase.Type = "binary"
			case 0:
				routeone, _ := strconv.Atoi(linearcaseregex.FindStringSubmatch(nextroutes[1])[1])
				appendcase.Routeone = routeone
				appendcase.Type = "linear"
			}
		case 1:
			appendcase.Body = strings.Split(nextroutes[0], "break;")[0]
			objin.Escapes = append(objin.Escapes, caseindex)
		}
		objin.Cases[caseindex] = appendcase
	}
	return objin
}
func switchFlattener(scriptin string) string {
	switchindices := switchregex.FindAllStringIndex(scriptin, -1)
	if len(switchindices) == 0{
		return scriptin
	}
	listin := []*switchObj{}
	for _, switchindex := range switchindices{
		switchend := sliceFunction(scriptin, switchindex[0])
		switchscript := scriptin[switchindex[0]:switchend]
		switchvar := switchregex.FindStringSubmatch(switchscript)
		init, _ := strconv.Atoi(switchvar[2])
		breakcase, _ := strconv.Atoi(strings.TrimSpace(switchvar[3]))
		switchobject := switchObj{
			Variable: switchvar[1],
			Initialization: init,
			Escapes: []int{breakcase},
			Cases: make(map[int]caseObj),
			Startindex: switchindex[0],
			Endindex: switchend,
			Basescript: switchscript,
		}
		listin = append(listin, &switchobject)
	}
	stackin := []*switchObj{
		{
			Startindex: 0,
			Endindex: len(scriptin),
			Basescript: scriptin,
		},
	}
	for index := 0; index < len(listin); index++{
		for listin[index].Startindex > stackin[len(stackin)-1].Endindex{
			stackin = stackin[:len(stackin)-1]
		}
		stackin[len(stackin)-1].Basescript = strings.Replace(stackin[len(stackin)-1].Basescript, listin[index].Basescript, fmt.Sprintf("subswitch%d!",index), -1)
		stackin = append(stackin, listin[index])
	}
	scriptin = stackin[0].Basescript
	for i, swobj := range listin{
		casedobj := caseCreator(*swobj)
		repval := switchwalker(casedobj.Initialization, casedobj.Escapes, casedobj.Cases)
		scriptin = strings.Replace(scriptin, fmt.Sprintf("subswitch%d!", i), repval, -1)
	}
	return scriptin
}
func getPosition(str, subString string, index int)  int{
	return len(strings.Join(strings.SplitN(str, subString, index+1)[:index], subString))
}
func sliceFunction(scriptin string, beginindex int) int {
	scriptout := scriptin
	escapecount := 1
	endindex := beginindex + strings.Index(scriptin[beginindex:],"{")+1
	for escapecount != 0 {
		switch string(scriptout[endindex]) {
		case "{":
			endindex++
			escapecount++
			break
		case "}":
			endindex++
			escapecount--
			break
		default:
			endindex++
			break
		}
	}
	return endindex
}
func boolArray(scriptin string) map[int]map[int]int {
	arrayinitvar := boolArrayInitialization.FindStringSubmatch(scriptin)
	boolarraymatches := []int{}
	for _, val := range arrayinitvar {
		num, _ := strconv.Atoi(val)
		boolarraymatches = append(boolarraymatches, num)
	}
	retmap := make(map[int]map[int]int)
	for foo := 0; foo < boolarraymatches[1]; foo++ {
		retmap[foo] = make(map[int]int)
	}
	for iterator1 := 0; iterator1 < boolarraymatches[1]; iterator1++ {
		for iterator2 := boolarraymatches[1] - 1; iterator2 >= 0; iterator2-- {
			ind1 := 0
			ind2 := boolarraymatches[3]
			ind3 := boolarraymatches[3]
			if (iterator2 >= ind2) {
				ind1 = ind2
				ind2 = boolarraymatches[4]
				ind3 = boolarraymatches[4] - boolarraymatches[3]
			}
			retmap[iterator1][ind1+(iterator2-ind1+(boolarraymatches[2]*iterator1))%(ind3)] = iterator2
		}
	}
	return retmap
}
func boolArrayEvaluate(boolstring string, boolarrin map[int]map[int]int) int {
	firstindexes := []int{}
	for _, val := range digitmatch.FindAllString(multipleindexmatcher.FindString(boolstring), -1) {
		appval, _ := strconv.Atoi(val)
		firstindexes = append(firstindexes, appval)
	}
	firstval := boolarrin[firstindexes[0]][firstindexes[1]]
	if len(firstindexes) > 2 {
		for _, val := range firstindexes[2:] {
			firstval = boolarrin[firstval][val]
		}
	}
	return firstval
}
func replaceBoolArrays(scriptin string, boolarrin map[int]map[int]int) string{
	for _, boolarrstring := range boolarrmatch.FindAllString(scriptin, -1){
		numstring := strconv.Itoa(boolArrayEvaluate(boolarrstring, boolarrin))
		scriptin = strings.Replace(scriptin, boolarrstring, numstring, 1)
	}
	return scriptin
}

func convertToInt32(roundsin []string)[]string{
	retarr := []string{}
	for i := 0; i < len(roundsin)/4; i++{
		roundnum := 0
		matchval1 := keydigitmatch.FindStringSubmatch(roundsin[4*i])
		val1, _ := strconv.Atoi(matchval1[1]+matchval1[2])
		roundnum += convertToJsNum(val1 << 24)
		matchval2 := keydigitmatch.FindStringSubmatch(roundsin[(4*i)+1])
		val2, _ := strconv.Atoi(matchval2[1]+matchval2[2])
		roundnum += val2 << 16
		matchval3 := keydigitmatch.FindStringSubmatch(roundsin[(4*i)+2])
		val3, _ := strconv.Atoi(matchval3[1]+matchval3[2])
		roundnum += val3 << 8
		matchval4 := keydigitmatch.FindStringSubmatch(roundsin[(4*i)+3])
		val4, _ := strconv.Atoi(matchval4[1]+matchval4[2])
		roundnum += val4
		retarr = append(retarr, strconv.Itoa(roundnum))
	}
	return retarr
}
func convertToJsNum(intin int)int{
	return int(int32(intin))
}
func o5f(decodestringin string, f6f, Q6f, index int) string {
	var C6f strings.Builder
	indexer1 := 0
	stringarray := []string{strconv.Itoa(f6f)}
	stringarraylen := len(stringarray[indexer1])
	onelessdecodestring := len(decodestringin) - 1
	indexer2 := 0
	for onelessdecodestring >= 0 {
		if indexer2 == stringarraylen {
			indexer2 = 0
			indexer1++
			if indexer1 == Q6f {
				indexer1 = 0
			}
			if len(stringarray) < Q6f {
				stringarray = append(stringarray, strconv.Itoa(c5f(stringarray[indexer1-1], stringarray[indexer1-1])))
			}
			stringarraylen = len(stringarray[indexer1])
		}
		C6f.WriteString(fromCharCode(charCodeAt(decodestringin,onelessdecodestring)^ charCodeAt(stringarray[indexer1],indexer2)))
		onelessdecodestring--
		indexer2++
	}
	return Reverse(C6f.String())
}
func g5f(B5f string, i5f, F5f int) int {
	P5f := 0xcc9e2d51
	T5f := 0x1b873593
	var R5f int
	G5f := F5f
	z5f := i5f
	z5f = z5f & ^0x3
	K5f := 0
	for K5f < z5f {
		R5f = charCodeAt(B5f, K5f)&0xff | (charCodeAt(B5f, K5f+1)&0xff)<<8 | (charCodeAt(B5f, K5f+2)&0xff)<<16 | (charCodeAt(B5f, K5f+3)&0xff)<<24
		R5f = r5f(R5f, P5f)
		R5f = convertToJsNum((R5f&0x1ffff)<<15 | int(uint32(R5f)>>17))
		R5f = r5f(R5f, T5f)
		G5f ^= R5f
		G5f = convertToJsNum((G5f&0x7ffff)<<13 | int(uint32(G5f)>>19))
		G5f = convertToJsNum(G5f*5 + 0xe6546b64 | 0)
		K5f += 4
	}
	R5f = 0
	I6f := i5f % 4
	switch I6f {
	case 3:
		R5f = (charCodeAt(B5f, z5f+2) & 0xff) << 16
		R5f |= (charCodeAt(B5f, z5f+1) & 0xff) << 8
		R5f |= charCodeAt(B5f, z5f) & 0xff
		R5f = r5f(R5f, P5f)
		R5f = convertToJsNum((R5f&0x1ffff)<<15 | int(uint32(R5f)>>17))
		R5f = r5f(R5f, T5f)
		G5f ^= R5f
		G5f ^= i5f
		G5f ^= int(uint32(G5f) >> 16)
		G5f = r5f(G5f, 0x85ebca6b)
		G5f ^= int(uint32(G5f) >> 13)
		G5f = r5f(G5f, 0xc2b2ae35)
		G5f ^= int(uint32(G5f) >> 16)
		return G5f
	case 2:
		R5f |= (charCodeAt(B5f, z5f+1) & 0xff) << 8
		R5f |= charCodeAt(B5f, z5f) & 0xff
		R5f = r5f(R5f, P5f)
		R5f = convertToJsNum((R5f&0x1ffff)<<15 | int(uint32(R5f)>>17))
		R5f = r5f(R5f, T5f)
		G5f ^= R5f
		G5f ^= i5f
		G5f ^= int(uint32(G5f) >> 16)
		G5f = r5f(G5f, 0x85ebca6b)
		G5f ^= int(uint32(G5f) >> 13)
		G5f = r5f(G5f, 0xc2b2ae35)
		G5f ^= int(uint32(G5f) >> 16)
		return G5f
	case 1:
		R5f |= charCodeAt(B5f, z5f) & 0xff
		R5f = r5f(R5f, P5f)
		R5f = convertToJsNum((R5f&0x1ffff)<<15 | int(uint32(R5f)>>17))
		R5f = r5f(R5f, T5f)
		G5f ^= R5f
		G5f ^= i5f
		G5f ^= int(uint32(G5f) >> 16)
		G5f = r5f(G5f, 0x85ebca6b)
		G5f ^= int(uint32(G5f) >> 13)
		G5f = r5f(G5f, 0xc2b2ae35)
		G5f ^= int(uint32(G5f) >> 16)
		return G5f
	default:
		G5f ^= i5f
		G5f ^= int(uint32(G5f) >> 16)
		G5f = r5f(G5f, 0x85ebca6b)
		G5f ^= int(uint32(G5f) >> 13)
		G5f = r5f(G5f, 0xc2b2ae35)
		G5f ^= int(uint32(G5f) >> 16)
		return G5f
	}
}
func c5f(strargsin ...string) int {
	I5f := strargsin[0]
	if len(strargsin) == 1 {
		I5f = strings.Join(A5f.Split(I5f, 2), "$1$3")
		I5f = replacementregexp.ReplaceAllString(I5f, "")
		return g5f(I5f, len(I5f), len(I5f))
	} else {
		I5f = replacementregexp.ReplaceAllString(I5f, "")
		return g5f(I5f, len(I5f), len(I5f))
	}
}
func r5f(x5f, S5f int) int {
	M5f := S5f & 0xffff
	y5f := S5f - M5f
	return convertToJsNum(convertToJsNum(y5f*x5f) + convertToJsNum(M5f * x5f))
}
func N5f(e5f string) string {
	e5f = strings.ReplaceAll(e5f, "(", "")
	e5f = strings.ReplaceAll(e5f, ")", "")
	return e5f
}
func helper(rawfuncstr, decodestr string, index int) string {
	decoded, _ := url.PathUnescape(decodestr)
	return o5f(decoded, c5f(N5f(rawfuncstr)), 5, index)
}
func decodeStrings(scriptin string) []string {
	var retarr []string
	obfuscatedstrings := obffuncregex.FindAllStringSubmatch(scriptin, -1)
	for i, strarr := range obfuscatedstrings{
		result := helper(strarr[1], strarr[2], i)
		if strings.Contains(result, "class"){
			retarr = append(retarr, result)
		}
	}
	return retarr
}
func globalvarmapping(scriptin string) map[string]string{
	retmap := make(map[string]string)
	for _, match := range basicvardef1.FindAllStringSubmatch(scriptin, -1){
		retmap[match[1]] = match[2]
	}
	for _, match := range basicvardef2.FindAllStringSubmatch(scriptin, -1){
		retmap[match[1]] = match[2]
	}
	return retmap
}
func globalvarreplacer(bodyin string, varmap map[string]string) string{
	for _, match := range basicvarmatch1.FindAllStringSubmatch(bodyin, -1){
		bodyin = strings.Replace(bodyin, match[1], varmap[match[1]], -1)
	}
	for _, match := range basicvarmatch2.FindAllStringSubmatch(bodyin, -1){
		bodyin = strings.Replace(bodyin, match[1], varmap[match[1]], -1)
	}
	return bodyin
}

func flagfinder(scriptin, basescript string, varmap map[string]string, boolarrayin map[int]map[int]int, stringarrin []string) []string{
	retkeys := []string{}
	keys := []string{}
	bodies := []string{}
	nextbegin := strings.Index(scriptin, "{")
	nextend := 0
	prior := 0
	for nextbegin > 0{
		prior = nextend+nextbegin
		keys = append(keys, scriptin[nextend:prior])
		nextend = sliceFunction(scriptin, prior)
		bodies = append(bodies, scriptin[prior:nextend])
		nextbegin = strings.Index(scriptin[nextend:], "{")
	}
	switch len(keys){
	case 5:
		bodies[2] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[2], varmap), boolarrayin))
		bodies[3] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[3], varmap), boolarrayin))
		for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[2], varmap))[2:6]{
			if strings.Contains(submatch, "["){
				for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
					interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
					matches := interestdef.FindAllStringSubmatch(bodies[2],-1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(scriptin, -1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(basescript, -1)
							if len(matches) == 0{
								basescript = globalvarreplacer(basescript, varmap)
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
							}
						}
					}
					submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
				}
			}
			retkeys = append(retkeys, digitmatch.FindString(submatch))
		}
		for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[3], varmap))[2:6]{
			if strings.Contains(submatch, "["){
				for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
					interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
					matches := interestdef.FindAllStringSubmatch(bodies[3],-1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(scriptin, -1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(basescript, -1)
							if len(matches) == 0{
								basescript = globalvarreplacer(basescript, varmap)
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
							}
						}
					}
					submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
				}
			}
			retkeys = append(retkeys, digitmatch.FindString(submatch))
		}
	case 6:
		if strings.Contains(keys[2], "static"){
			bodies[3] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[3], varmap), boolarrayin))
			bodies[4] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[4], varmap), boolarrayin))
			for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[3], varmap))[2:6]{
				if strings.Contains(submatch, "["){
					for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
						interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
						matches := interestdef.FindAllStringSubmatch(bodies[3],-1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(scriptin, -1)
							if len(matches) == 0{
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
								if len(matches) == 0{
									basescript = globalvarreplacer(basescript, varmap)
									matches = interestdef.FindAllStringSubmatch(basescript, -1)
								}
							}
						}
						submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
					}
				}
				retkeys = append(retkeys, digitmatch.FindString(submatch))
			}
			for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[4], varmap))[2:6]{
				if strings.Contains(submatch, "["){
					for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
						interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
						matches := interestdef.FindAllStringSubmatch(bodies[4],-1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(scriptin, -1)
							if len(matches) == 0{
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
								if len(matches) == 0{
									basescript = globalvarreplacer(basescript, varmap)
									matches = interestdef.FindAllStringSubmatch(basescript, -1)
								}
							}
						}
						submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
					}
				}
				retkeys = append(retkeys, digitmatch.FindString(submatch))
			}
		}else if strings.Contains(keys[4], "()"){
			bodies[2] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[2], varmap), boolarrayin))
			bodies[4] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[4], varmap), boolarrayin))
			for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[2], varmap))[2:6]{
				if strings.Contains(submatch, "["){
					for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
						interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
						matches := interestdef.FindAllStringSubmatch(bodies[2],-1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(scriptin, -1)
							if len(matches) == 0{
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
								if len(matches) == 0{
									basescript = globalvarreplacer(basescript, varmap)
									matches = interestdef.FindAllStringSubmatch(basescript, -1)
								}
							}
						}
						submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
					}
				}
				retkeys = append(retkeys, digitmatch.FindString(submatch))
			}
			for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[4], varmap))[2:6]{
				if strings.Contains(submatch, "["){
					for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
						interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
						matches := interestdef.FindAllStringSubmatch(bodies[4],-1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(scriptin, -1)
							if len(matches) == 0{
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
								if len(matches) == 0{
									basescript = globalvarreplacer(basescript, varmap)
									matches = interestdef.FindAllStringSubmatch(basescript, -1)
								}
							}
						}
						submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
					}
				}
				retkeys = append(retkeys, digitmatch.FindString(submatch))
			}
		}else {
			bodies[3] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[3], varmap), boolarrayin))
			bodies[4] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[4], varmap), boolarrayin))
			challengevar := webglchallengematch.FindStringSubmatch(bodies[4])[2]
			challengestring := ""
			for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[3], varmap))[2:6]{
				if strings.Contains(submatch, "["){
					for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
						interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
						matches := interestdef.FindAllStringSubmatch(bodies[3],-1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(scriptin, -1)
							if len(matches) == 0{
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
								if len(matches) == 0{
									basescript = globalvarreplacer(basescript, varmap)
									matches = interestdef.FindAllStringSubmatch(basescript, -1)
								}
							}
						}
						submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
					}
				}
				retkeys = append(retkeys, digitmatch.FindString(submatch))
			}
			for _, submatch := range regexp.MustCompile(fmt.Sprintf("%s\\+?=\\w{3,4}\\.\\w{3}\\((.*?)\\);", strings.Replace(challengevar,"[","\\[",-1))).FindAllStringSubmatch(bodies[4],-1){
				if strings.Contains(submatch[1], "["){
					for _, bracketvar := range interestmatch.FindAllString(submatch[1], -1){
						interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
						matches := interestdef.FindAllStringSubmatch(bodies[4],-1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(scriptin, -1)
							if len(matches) == 0{
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
								if len(matches) == 0{
									basescript = globalvarreplacer(basescript, varmap)
									matches = interestdef.FindAllStringSubmatch(basescript, -1)
								}
							}
						}
						submatch[1] = strings.Replace(submatch[1], bracketvar, matches[len(matches)-1][1],-1)
					}
				}
				strarrindex, _ := strconv.Atoi(digitmatch.FindString(submatch[1]))
				challengestring += stringarrin[strarrindex]
			}
			challengepush := 0
			for i, floatvar := range webglfloats.FindAllStringSubmatch(challengestring,-1){
				floatval, _ := strconv.ParseFloat(floatvar[1], 64)
				if (i+1)%2 == 0{
					challengepush += bdfh[int(floatval/0.0125)]
					retkeys = append(retkeys, strconv.Itoa(challengepush))
					challengepush = 0
				}else{
					challengepush += aceg[int(floatval/0.0125)]
				}
			}
		}
	case 7:
		bodies[4] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[4], varmap), boolarrayin))
		bodies[5] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[5], varmap), boolarrayin))
		for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[4], varmap))[2:6]{
			if strings.Contains(submatch, "["){
				for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
					interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
					matches := interestdef.FindAllStringSubmatch(bodies[4],-1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(scriptin, -1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(basescript, -1)
							if len(matches) == 0{
								basescript = globalvarreplacer(basescript, varmap)
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
							}
						}
					}
					submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
				}
			}
			retkeys = append(retkeys, digitmatch.FindString(submatch))
		}
		for _, submatch := range pushregex.FindStringSubmatch(globalvarreplacer(bodies[5], varmap))[2:6]{
			if strings.Contains(submatch, "["){
				for _, bracketvar := range interestmatch.FindAllString(submatch, -1){
					interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
					matches := interestdef.FindAllStringSubmatch(bodies[5],-1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(scriptin, -1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(basescript, -1)
							if len(matches) == 0{
								basescript = globalvarreplacer(basescript, varmap)
								matches = interestdef.FindAllStringSubmatch(basescript, -1)
							}
						}
					}
					submatch = strings.Replace(submatch, bracketvar, matches[len(matches)-1][1],-1)
				}
			}
			retkeys = append(retkeys, digitmatch.FindString(submatch))
		}
	default:
		return retkeys
	}
	return retkeys
}
func keyfinder(scriptin, rawstring string, varmap map[string]string, boolarrayin map[int]map[int]int) ([][]string,[][]string){
	keys := []string{}
	bodies := []string{}
	nextbegin := strings.Index(scriptin, "{")
	nextend := 0
	prior := 0
	for nextbegin > 0{
		prior = nextend+nextbegin
		keys = append(keys, scriptin[nextend:prior])
		nextend = sliceFunction(scriptin, prior)
		bodies = append(bodies, scriptin[prior:nextend])
		nextbegin = strings.Index(scriptin[nextend:], "{")
	}
	bodies[0] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[0], varmap), boolarrayin))
	bodies[1] = switchFlattener(replaceBoolArrays(globalvarreplacer(bodies[1], varmap), boolarrayin))

	enckey := [][]string{}
	encroundkeys := keymatch.FindAllStringSubmatch(bodies[0], -1)
	encroundkeyset := []string{}
	for _, key := range encroundkeys{
		if len(encroundkeyset) == 0{
			encroundkeyset = append(encroundkeyset, key[1])
		}else{
			for _, appkey := range encroundkeyset{
				if key[1] == appkey{
					continue
				}else{
					encroundkeyset = append(encroundkeyset, key[1])
				}
			}
		}
	}
	firstround := strings.Split(bodies[0][strings.Index(bodies[0],encroundkeyset[0]):strings.Index(bodies[0],"return new")], encroundkeyset[0])
	appround := []string{}
	for _, round := range firstround[len(firstround)-4:]{
		appval := round[strings.LastIndex(round,",")+1:strings.LastIndex(round,")")]
		if strings.Contains(appval, "["){
			for _, bracketvar := range interestmatch.FindAllString(appval, -1){
				interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
				matches := interestdef.FindAllStringSubmatch(bodies[0],-1)
				if len(matches) == 0{
					matches = interestdef.FindAllStringSubmatch(scriptin, -1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						if len(matches) == 0{
							rawstring = globalvarreplacer(rawstring, varmap)
							matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						}
					}
				}
				appval = strings.Replace(appval, bracketvar, matches[len(matches)-1][1],-1)
			}
		}
		appendvalue := keydigitmatch.FindStringSubmatch(appval)
		appround = append(appround, appendvalue[1]+appendvalue[2])
	}
	enckey = append(enckey, appround)
	appround = []string{}
	midrounds := strings.Split(strings.Split(bodies[0][strings.Index(bodies[0], encroundkeyset[1]):strings.LastIndex(bodies[0], encroundkeyset[1])], "=[];")[1], fmt.Sprintf("=%s",encroundkeyset[1]))
	for _, block := range midrounds{
		rounds := strings.Split(block, encroundkeyset[1])[1:5]
		for _, round := range rounds{
			appval := round[strings.LastIndex(round,",")+1:strings.LastIndex(round,");")]
			if strings.Contains(appval, "["){
				for _, bracketvar := range interestmatch.FindAllString(appval, -1){
					interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
					matches := interestdef.FindAllStringSubmatch(bodies[0],-1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(scriptin, -1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(rawstring, -1)
							if len(matches) == 0{
								rawstring = globalvarreplacer(rawstring, varmap)
								matches = interestdef.FindAllStringSubmatch(rawstring, -1)
							}
						}
					}
					appval = strings.Replace(appval, bracketvar, matches[len(matches)-1][1],-1)
				}
			}
			appendvalue := keydigitmatch.FindStringSubmatch(appval)
			appround = append(appround, appendvalue[1]+appendvalue[2])
		}
		enckey = append(enckey, appround)
		appround = []string{}
	}
	finalroundkey := retvar.FindStringSubmatch(bodies[0])[1]
	finalrounds := strings.Split(bodies[0], finalroundkey)
	for _, fround := range finalrounds[len(finalrounds)-17:len(finalrounds)-1]{
		appval := fround[strings.LastIndex(fround,",")+1:strings.LastIndex(fround,");")]
		if strings.Contains(appval, "["){
			for _, bracketvar := range interestmatch.FindAllString(appval, -1){
				interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
				matches := interestdef.FindAllStringSubmatch(bodies[0],-1)
				if len(matches) == 0{
					matches = interestdef.FindAllStringSubmatch(scriptin, -1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						if len(matches) == 0{
							rawstring = globalvarreplacer(rawstring, varmap)
							matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						}
					}
				}
				appval = strings.Replace(appval, bracketvar, matches[len(matches)-1][1],-1)
			}
		}
		appendvalue := keydigitmatch.FindStringSubmatch(appval)
		appround = append(appround, appendvalue[1]+appendvalue[2])
	}
	appround = convertToInt32(appround)
	enckey = append(enckey, appround)

	deckey := [][]string{}
	encroundkeys = keymatch.FindAllStringSubmatch(bodies[1], -1)
	encroundkeyset = []string{}
	for _, key := range encroundkeys{
		if len(encroundkeyset) == 0{
			encroundkeyset = append(encroundkeyset, key[1])
		}else{
			for _, appkey := range encroundkeyset{
				if key[1] == appkey{
					continue
				}else{
					encroundkeyset = append(encroundkeyset, key[1])
				}
			}
		}
	}
	firstround = strings.Split(bodies[1][strings.Index(bodies[1],encroundkeyset[0]):strings.Index(bodies[1],"return new")], encroundkeyset[0])
	appround = []string{}
	for _, round := range firstround[len(firstround)-4:]{
		appval := round[strings.LastIndex(round,",")+1:strings.LastIndex(round,")")]
		if strings.Contains(appval, "["){
			for _, bracketvar := range interestmatch.FindAllString(appval, -1){
				interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
				matches := interestdef.FindAllStringSubmatch(bodies[1],-1)
				if len(matches) == 0{
					matches = interestdef.FindAllStringSubmatch(scriptin, -1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						if len(matches) == 0{
							rawstring = globalvarreplacer(rawstring, varmap)
							matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						}
					}
				}
				appval = strings.Replace(appval, bracketvar, matches[len(matches)-1][1],-1)
			}
		}
		appendvalue := keydigitmatch.FindStringSubmatch(appval)
		appround = append(appround, appendvalue[1]+appendvalue[2])
	}
	deckey = append(deckey, appround)
	appround = []string{}
	midrounds = strings.Split(strings.Split(bodies[1][strings.Index(bodies[1], encroundkeyset[1]):strings.LastIndex(bodies[1], encroundkeyset[1])], "=[];")[1], fmt.Sprintf("=%s",encroundkeyset[1]))
	for _, block := range midrounds{
		rounds := strings.Split(block, encroundkeyset[1])[1:5]
		for _, round := range rounds{
			appval := round[strings.LastIndex(round,",")+1:strings.LastIndex(round,");")]
			if strings.Contains(appval, "["){
				for _, bracketvar := range interestmatch.FindAllString(appval, -1){
					interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
					matches := interestdef.FindAllStringSubmatch(bodies[1],-1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(scriptin, -1)
						if len(matches) == 0{
							matches = interestdef.FindAllStringSubmatch(rawstring, -1)
							if len(matches) == 0{
								rawstring = globalvarreplacer(rawstring, varmap)
								matches = interestdef.FindAllStringSubmatch(rawstring, -1)
							}
						}
					}
					appval = strings.Replace(appval, bracketvar, matches[len(matches)-1][1],-1)
				}
			}
			appendvalue := keydigitmatch.FindStringSubmatch(appval)
			appround = append(appround, appendvalue[1]+appendvalue[2])
		}
		deckey = append(deckey, appround)
		appround = []string{}
	}
	keyarr := lastkeyregex.FindStringSubmatch(bodies[1])[1]
	lastroundmatch := regexp.MustCompile(fmt.Sprintf("%s\\[.{0,10}]=",strings.Replace(keyarr,"[","\\[",-1))).FindString(bodies[1])
	endscript := bodies[1][strings.Index(bodies[1],lastroundmatch):]
	finalrounds = strings.Split(endscript[0:strings.Index(endscript,"return new")], keyarr)
	for _, fround := range finalrounds[len(finalrounds)-16:]{
		appval := fround[strings.LastIndex(fround,",")+1:strings.LastIndex(fround,");")]
		if strings.Contains(appval, "["){
			for _, bracketvar := range interestmatch.FindAllString(appval, -1){
				interestdef := regexp.MustCompile(fmt.Sprintf("%s=\\+?\"?(\\w+)", strings.Replace(bracketvar, "[","\\[",-1)))
				matches := interestdef.FindAllStringSubmatch(bodies[1],-1)
				if len(matches) == 0{
					matches = interestdef.FindAllStringSubmatch(scriptin, -1)
					if len(matches) == 0{
						matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						if len(matches) == 0{
							rawstring = globalvarreplacer(rawstring, varmap)
							matches = interestdef.FindAllStringSubmatch(rawstring, -1)
						}
					}
				}
				appval = strings.Replace(appval, bracketvar, matches[len(matches)-1][1],-1)
			}
		}
		appendvalue := keydigitmatch.FindStringSubmatch(appval)
		appround = append(appround, appendvalue[1]+appendvalue[2])
	}
	appround = convertToInt32(appround)
	deckey = append(deckey, appround)
	return enckey, deckey
}
func mainstringarrresolve(scriptin string) []string {
	var D9j string
	e9j := stringarrregex.FindStringSubmatch(scriptin)[1]
	A9j, _ := url.PathUnescape(returnR.FindStringSubmatch(scriptin)[1])
	var g9j, z9j int
	for g9j < len(A9j) {
		if z9j == len(e9j) {
			z9j = 0
		}
		D9j += fromCharCode(charCodeAt(A9j, g9j) ^ charCodeAt(e9j, z9j))
		g9j++
		z9j++
	}
	arrout :=  strings.Split(D9j, splitR.FindStringSubmatch(scriptin)[1])

	iifestart := getPosition(scriptin, "(function", 3)
	iifeend := sliceFunction(scriptin, iifestart)
	iifefunc := scriptin[iifestart:iifeend]
	shiftscript := switchFlattener(iifefunc)
	shiftarr := shiftR.FindAllStringSubmatch(shiftscript, -1)
	shifterhold := []int{}
	for _, submatch := range shiftarr{
		val1, _ := strconv.Atoi(submatch[1])
		val2, _ := strconv.Atoi(submatch[2])
		shifterhold = append(shifterhold, val1, val2)
		if len(shifterhold) == 4{
			si1 := (len(arrout)+shifterhold[0])%len(arrout)
			si2 := (len(arrout)+shifterhold[1])%len(arrout)
			splicedarray := arrout[si1:si1+si2]
			si3 := (len(splicedarray)+shifterhold[2])%len(splicedarray)
			si4 := (len(splicedarray)+shifterhold[3])%len(splicedarray)
			reappend := splicedarray[si3:si3+si4]
			arrout = append(append(reappend, arrout[0:si1]...), arrout[si1+si2:]...)
			shifterhold = []int{}
		}
	}
	return arrout
}


func stringarrtointnested(arrin [][]string) [][]int32{
	retarr := [][]int32{}
	for _, encround := range arrin{
		appround := []int32{}
		for _, numstring := range encround{
			appval, _ := strconv.Atoi(numstring)
			appround = append(appround, int32(appval))
		}
		retarr = append(retarr, appround)
	}
	return retarr
}
func stringarrtoint(arrin []string) []int{
	retarr := []int{}
	for _, numstring := range arrin{
		appval, _ := strconv.Atoi(numstring)
		retarr = append(retarr, appval)
	}
	return retarr
}
func fromBytes(bytes []byte) string{
	var result []string
	for i := 0; i < len(bytes); i++{
		var v = bytes[i]
		result = append(result, string(Hex[(v & 0xf0) >> 4] + Hex[v & 0x0f]))
	}
	return strings.Join(result, "")
}
func toBytes(text string) []byte{
	reply, _ := hex.DecodeString(text)
	return reply
}
func Reverse(s string) string{
	size := len(s)
	buf := make([]byte, size)
	for start :=0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}
func fromCharCode(code int) string {
	return string(code)
}
func charCodeAt(s string, n int) int {
	return int(s[n])
}

func intsliceincludes(slicein []int, matchint  int) bool{
	for _, ind := range slicein{
		if ind == matchint{
			return true
		}
	}
	return false
}