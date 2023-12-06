package funcaptcha

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func getF() string {
	var res []string
	for _, val := range fe {
		for _, v := range val {
			res = append(res, fmt.Sprintf("%v", v))
		}
	}
	return getMurmur128String(strings.Join(res, "~~~"), 31)
}

func getN() string {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/1000000000)
	return base64.StdEncoding.EncodeToString([]byte(timestamp))
}

func getWh() string {
	return fmt.Sprintf("%s|%s", getWindowHash(), getWindowProtoChainHash())
}

func getFe() string {
	fe, _ := json.Marshal(getFeList())
	return string(fe)
}

func getFeList() []string {
	// var b6 = [];
	var feList []string
	for _, feMap := range fe {
		for k, v := range feMap {
			if k == "S" ||
				k == "AS" ||
				k == "JSF" ||
				k == "T" {
				v = strings.ReplaceAll(v.(string), ";", ",")
			} else if k == "CFP" { // case dH(f_a_iI.X):
				v = getCFPHash(cfp)
			} else if k == "P" { // case 'P':
				v = getP(p)
			}
			feList = append(feList, fmt.Sprintf("%v:%v", k, v))
		}
	}
	return feList
}

func getP(p string) string {
	var pList []string
	for _, s := range strings.Split(p, ";") {
		split := strings.Split(s, "::")
		pList = append(pList, split[0])
	}
	return strings.Join(pList, ",")
}
