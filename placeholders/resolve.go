package placeholders

import (
	"regexp"
	"strings"
	"fmt"
)

var PlaceHoldersRegExp = regexp.MustCompile(`\{(.*?)\}`)

// resolvePlaceholders resolves placeholders of s by vars
func Resolve(s string, vars map[string]interface{}) (interface{}, error) {
	matches := PlaceHoldersRegExp.FindAllStringSubmatch(s, -1)
	if matches == nil {
		// no matches found
		return s, nil
	}

	if len(matches) == 1 && strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		ph := matches[0][1]
		// single place holder, can be resolved to any type
		if val, ok := vars[ph]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("error resolving var [%s]", ph)
	}

	for _, m := range matches {
		ph := m[0]
		val := fmt.Sprintf("%s", m[1])
		resolvedVal := vars[val]
		if resolvedVal == nil {
			return s, fmt.Errorf("var [%s] does not exist", val)
		}
		s = strings.Replace(s, ph, fmt.Sprintf("%v", resolvedVal), -1)
	}
	return s, nil
}
