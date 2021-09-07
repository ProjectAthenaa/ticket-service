package deob

import (
	"github.com/ProjectAthenaa/ticket-service/models"
	"strings"
)

func GetVersion(scriptin, hash string) *models.Version {
	enckey := [][]string{}
	deckey := [][]string{}
	userflags := []string{}
	boolarr := boolArray(scriptin)
	stringarray := mainstringarrresolve(scriptin)
	varmap := globalvarmapping(scriptin)
	var cparam string
	for i, subfunc := range decodeStrings(scriptin) {
		startindex := strings.Index(subfunc, "class")
		body := subfunc[strings.Index(subfunc[startindex:], "{")+startindex+1 : sliceFunction(subfunc, startindex)-1]
		staticcount := len(staticmatch.FindAllString(body, -1))
		if i == 0 {
			roundvarmap := map[string]string{}
			for _, val := range threeletterdef.FindAllStringSubmatch(subfunc, -1) {
				roundvarmap[val[1]] = val[2]
			}
			for _, val := range threeletterarr.FindAllStringSubmatch(subfunc, -1) {
				scriptin = strings.Replace(subfunc, val[0], "["+roundvarmap[val[1]]+"]", 1)
			}
			subfunc = switchFlattener(replaceBoolArrays(globalvarreplacer(subfunc, varmap), boolarr))
			for _, val := range varindexset.FindAllStringSubmatch(subfunc, -1) {
				roundvarmap[val[1]] = val[2]
			}
			for _, val := range stringfuncresolve.FindAllStringSubmatch(subfunc, -1) {
				indexval := varindex.FindStringSubmatch(val[1])
				if len(indexval) > 0 {
					body = strings.Replace(body, indexval[0], roundvarmap[indexval[0]], -1)
				}
			}
			cparam = cparamresolver(body, varmap, boolarr, stringarray)
		}
		if staticcount < 3 {
			userflags = append(userflags, flagfinder(body, subfunc, varmap, boolarr, stringarray)...)
		} else if staticcount > 9 {
			enckey, deckey = keyfinder(body, subfunc, varmap, boolarr)
		}
	}

	version := &models.Version{
		EncKeys: stringarrtointnested(enckey),
		DecKeys: stringarrtointnested(deckey),
		Flags:   stringarrtoint(userflags),
		CParam:  cparam,
		Hash:    hash,
	}

	return version
}
