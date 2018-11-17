package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//
// Types
//
type chanChan chan chan struct{}

//
// Global variables
//

// Defaults
var defCategoryID = 22
var defPrivacyStatus = "public"
var defLicense = ""

// Regexps
var reOpen = regexp.MustCompile("(?i)open|free|public|common|creative")

// Category
var categoryName = map[string]int{
	"Film & Animation ":     1,
	"Autos & Vehicles":      2,
	"Music":                 10,
	"Pets & Animals":        15,
	"Sports":                17,
	"Short Movies":          18,
	"Travel & Events":       19,
	"Gaming":                20,
	"Videoblogging":         21,
	"People & Blogs":        22,
	"Comedy":                23,
	"Entertainment":         24,
	"News & Politics":       25,
	"Howto & Style":         26,
	"Education":             27,
	"Science & Technology":  28,
	"Nonprofits & Activism": 29,
	"Movies":                30,
	"Anime/Animation":       31,
	"Action/Adventure":      32,
	"Classics":              33,
	"Documentary":           35,
	"Drama":                 36,
	"Family":                37,
	"Foreign":               38,
	"Horror":                39,
	"Sci-Fi/Fantasy":        40,
	"Thriller":              41,
	"Shorts":                42,
	"Shows":                 43,
	"Trailers":              44,
}
var categoryRe = map[*regexp.Regexp]int{}

//
// Functions
//

func regexpIntMap(out map[*regexp.Regexp]int, inp map[string]int) {
	sep := regexp.MustCompile("\\W+")
	for k, v := range inp {
		ka := strings.ToLower(sep.ReplaceAllString(k, "|"))
		out[regexp.MustCompile(ka)] = v
	}
}

func parseBool(txt string, def bool) bool {
	ans, err := strconv.ParseBool(txt)
	if err != nil {
		return def
	}
	return ans
}

func parseInt(txt string, def int) int {
	ans, err := strconv.ParseInt(txt, 10, 32)
	if err != nil {
		return def
	}
	return int(ans)
}

func parseFloat(txt string, def float64) float64 {
	ans, err := strconv.ParseFloat(txt, 64)
	if err != nil {
		return def
	}
	return ans
}

func parseString(txt string, def string) string {
	if txt != "" {
		return txt
	}
	return def
}

func parseCategory(txt string) int {
	ans, err := strconv.ParseInt(txt, 10, 32)
	if err == nil {
		return int(ans)
	}
	for k, v := range categoryRe {
		if k.MatchString(txt) {
			return v
		}
	}
	return defCategoryID
}

func parsePrivacyStatus(txt string) string {
	if txt == "" || reOpen.FindString(txt) == "" {
		return defPrivacyStatus
	}
	return "private"
}

func parseLicense(txt string) string {
	if txt == "" || reOpen.FindString(txt) == "" {
		return defLicense
	}
	return "creativeCommon"
}

func shortString(txt string, max int) string {
	if len(txt) <= max {
		return txt
	}
	return txt[0:max-3] + "..."
}

func arrayJoin(arr []interface{}, sep string) string {
	var sb strings.Builder
	for _, val := range arr {
		sb.WriteString(fmt.Sprintf("%v%v", val, sep))
	}
	if sb.Len() == 0 {
		return ""
	}
	return sb.String()[0 : sb.Len()-len(sep)]
}

func mapGet(obj map[string]interface{}, pth string) interface{} {
	var dot = strings.Index(pth, ".")
	if dot < 0 {
		return obj[pth]
	}
	if val, ok := obj[pth[0:dot]].(map[string]interface{}); ok {
		return mapGet(val, pth[dot+1:])
	}
	return nil
}

func mapString(txt string, obj map[string]interface{}) string {
	return reTemplate.ReplaceAllStringFunc(txt, func(m string) string {
		var val = mapGet(obj, m[2:len(m)-1])
		if arr, ok := val.([]interface{}); ok {
			return arrayJoin(arr, ",")
		}
		return fmt.Sprintf("%v", val)
	})
}
