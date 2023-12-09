package utils

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "time"
)

var durationReg = regexp.MustCompile(`(\d+(\.\d+)?(/\d+(\.\d+)?)?)?([a-zA-Z]+)?`)

func parseFloats(value string) (float64, error) {
    if strings.Contains(value, "/") {
        parts := strings.Split(value, "/")
        if len(parts) == 2 {
            numerator, err := strconv.ParseFloat(parts[0], 64)
            if err != nil {
                return 0, err
            }
            denominator, err := strconv.ParseFloat(parts[1], 64)
            if err != nil {
                return 0, err
            }
            return numerator / denominator, nil
        }
        return 0, fmt.Errorf("invalid fraction format: %s", value)
    }

    return strconv.ParseFloat(value, 64)
}

func ProcessDuration(input string) (time.Duration, error) {
    matches := durationReg.FindAllStringSubmatch(input, -1)
    var total time.Duration
    for _, match := range matches {
        var value float64
        if match[1] != "" {
            var err error
            value, err = parseFloats(match[1])
            if err != nil {
                return 0, fmt.Errorf("%v", err)
            }
        }
        unit := strings.ToLower(match[5])
        switch unit {
        case "d":
            total += time.Duration(value * 24 * 60 * 60) * time.Second
        case "w":
            total += time.Duration(value * 7 * 24 * 60 * 60) * time.Second
        default:
            duration, err := time.ParseDuration(fmt.Sprintf("%f%s", value, unit))
            if err != nil {
                return 0, fmt.Errorf("%v", err)
            }
            total += duration
        }
    }
    return total, nil
}
