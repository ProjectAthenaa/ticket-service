package deob

import (
	"github.com/ProjectAthenaa/database-module/models"
	"strings"
)


func GetVersion(scriptin, agent string) models.TicketVersion{
	enckey := [][]string{}
	deckey := [][]string{}
	userflags := []string{}
	boolarr := boolArray(scriptin)
	stringarray := mainstringarrresolve(scriptin)
	varmap := globalvarmapping(scriptin)
	for _, subfunc := range decodeStrings(scriptin){
		startindex := strings.Index(subfunc, "class")
		body := subfunc[strings.Index(subfunc[startindex:],"{")+startindex+1: sliceFunction(subfunc, startindex)-1]
		staticcount := len(staticmatch.FindAllString(body,-1))
		if staticcount < 3{
			userflags = append(userflags, flagfinder(body, subfunc, varmap, boolarr, stringarray)...)
		}else if staticcount > 9{
			enckey, deckey = keyfinder(body, subfunc, varmap, boolarr)
		}
	}
	version := models.TicketVersion{
		Enc:   stringarrtointnested(enckey),
		Dec:   stringarrtointnested(deckey),
		Flag:  stringarrtoint(userflags),
		Agent: agent,
	}


	return version
}

