package types

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var prefixes = []string{"", "K", "M", "G", "T"}

func MarshalUnit(unit string, value float64) ([]byte, error) {
	order := 0
	var divided float64
	for divided = value; divided < 1000.0; divided = divided / 1000.0 {
		order++
	}

	if order > (len(prefixes) - 1) {
		return nil, fmt.Errorf("Unsupported frequency prefix")
	}

	str := fmt.Sprintf("%.2f %s%s", divided, prefixes[order], unit)

	return []byte(str), nil
}

var unitRegex = regexp.MustCompile(`^([0-9\.]+)[ ]{0,1}([a-zA-Z]+)$`)

func UnmarshalUnit(unit string, text []byte) (float64, error) {

	matches := unitRegex.FindStringSubmatch(string(text))
	if matches == nil {
		return 0.0, fmt.Errorf("Unit must be of the form 'Value PrefixUnit`, ie. '100.2 K%s'", unit)
	}

	valueString := matches[1]
	unitString := matches[2]

	// Check suffix matches
	if !strings.HasSuffix(unitString, unit) {
		return 0.0, fmt.Errorf("Unable to parse unit: '%s' expected suffix: '%s'", unitString, unit)
	}

	// Strip suffix
	order := 0
	prefix := strings.Replace(unitString, unit, "", -1)
	if prefix != "" {
		for i := range prefixes {
			if strings.ToLower(prefix) == strings.ToLower(prefixes[i]) {
				order = i
			}
		}
		if order == 0 {
			return 0.0, fmt.Errorf("Unrecognised SI prefix: '%s'", prefix)
		}
	}

	// Parse floating point component
	base, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0.0, err
	}

	// Multiply by prefix order
	value := base * math.Pow(1000, float64(order))

	return value, nil
}
