package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ConvertFromStringToIntMapMap(str string) (map[int]int, error) {
	if len(str) < 4 || str[:2] != "{{" || str[len(str)-2:] != "}}" {
		return nil, fmt.Errorf("invalid format: string must start with '{{' and end with '}}'")
	}

	content := str[2 : len(str)-2]
	if content == "" {
		return make(map[int]int), nil
	}

	pairs := strings.Split(content, "},{")
	result := make(map[int]int)

	for _, pair := range pairs {

		pair = strings.Trim(pair, "{}")
		kv := strings.Split(pair, ":")

		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid key-value pair format: %s", pair)
		}

		key, err := strconv.Atoi(strings.TrimSpace(kv[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid key: %s", kv[0])
		}

		value, err := strconv.Atoi(strings.TrimSpace(kv[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid value: %s", kv[1])
		}

		result[key] = value
	}

	return result, nil
}

func ConvertIntMapToString(m map[int]int) string {
	if len(m) == 0 {
		return "{{}}"
	}

	var pairs []string
	for key, value := range m {
		pairs = append(pairs, fmt.Sprintf("%d:%d", key, value))
	}

	return "{{" + strings.Join(pairs, "},{") + "}}"
}
