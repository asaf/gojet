package placeholders

import (
	"regexp"
	"strings"
	"fmt"
)

var PlaceHoldersRegExp = regexp.MustCompile(`\{(.*?)\}`)

// resolvePlaceholders resolves placeholders of s by vars
func Resolve(s string, vars map[string]interface{}, extras ...map[string]interface{}) (interface{}, error) {
	matches := PlaceHoldersRegExp.FindAllStringSubmatch(s, -1)
	if matches == nil {
		// no matches found
		return s, nil
	}

	if len(matches) == 1 && strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		ph := matches[0][1]

		// single place holder, can be resolved to any type
		v, err := resolveVar(ph, vars, extras...)
		if err != nil {
			return nil, err
		}

		return v, nil
	}

	for _, m := range matches {
		ph := m[1]

		resolvedVal, err := resolveVar(ph, vars, extras...)
		if err != nil {
			return nil, err
		}

		s = strings.Replace(s, m[0], fmt.Sprintf("%v", resolvedVal), -1)
	}
	return s, nil
}

func resolveVar(ph string, vars map[string]interface{}, extras ...map[string]interface{}) (interface{}, error) {
	// first try to resolve by vars
	if v, ok := vars[ph]; ok {
		return v, nil
	}

	// then try per extra map
	for _, extra := range extras {
		if v, ok := extra[ph]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("placeholder [%s] cannot be resolved", ph)
}
