package handlers

import (
	"strings"
	"unicode"
)

var platforms []string = []string{
	"Linux",
	"Android",
	"Windows",
	"Mac OS",
	"iPhone",
	"iPad",
	"PlayStation",
	"Xbox",
	"Macintosh",
}

func MassStrings(str string) []string {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && c == '(' || c == ')' || c == ';'
	}
	return strings.FieldsFunc(str, f)
}

func PLatfor(userAgentString string) string {
	var (
		//  переменная concatString нужна для поиска сложных ралтформ. Например LinuxAndroid - это Андроид
		concatString string
	)

	massString := MassStrings(userAgentString)

	for _, v := range massString {
		for _, m := range platforms {
			check := strings.Contains(v, m)
			if check {
				concatString += m
			}
		}
	}

	switch concatString {
	case "Linux":
		return "Linux"
	case "Android":
		return "Android"
	case "LinuxAndroid":
		return "Android"
	case "Windows":
		return "Windows"
	case "PlayStation":
		return "PlayStation"
	case "iPadMac OS":
		return "iPad"
	case "Macintosh":
		return "Macintosh"
	case "MacintoshMac OS":
		return "Mac OS"
	case "Mac OS":
		return "Mac OS"
	default:
		return "unknown"
	}
}
