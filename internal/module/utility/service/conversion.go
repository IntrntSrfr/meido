package service

import "fmt"

type ConversionType int

const (
	ConversionTypeTemperature ConversionType = iota
	ConversionTypeLength
	ConversionTypeWeight
	ConversionTypeVolume
	ConversionTypeTime
)

type Conversion struct {
	Name string
	InSI float64
}

var unitMappings = map[ConversionType]map[string]Conversion{
	ConversionTypeTemperature: {
		"c": {Name: "Celsius", InSI: 1.0},
		"f": {Name: "Fahrenheit", InSI: 1.0},
		"k": {Name: "Kelvin", InSI: 1.0},
	},
	ConversionTypeLength: {
		"cm": {Name: "Centimeter", InSI: 0.01},
		"m":  {Name: "Meter", InSI: 1.0},
		"km": {Name: "Kilometer", InSI: 1_000.0},
		"mm": {Name: "Millimeter", InSI: 0.001},
		"um": {Name: "Micrometer", InSI: 0.000001},
		"nm": {Name: "Nanometer", InSI: 0.000000001},
		"mi": {Name: "Mile", InSI: 1_609.34},
		"yd": {Name: "Yard", InSI: 0.9144},
		"ft": {Name: "Foot", InSI: 0.3048},
		"in": {Name: "Inch", InSI: 0.0254},
		"ly": {Name: "Light Year", InSI: 9_460_730_472_580.8},
		"au": {Name: "Astronomical Unit", InSI: 149_597_870_700.0},
		"pc": {Name: "Parsec", InSI: 30_856_775_814_671.7},
	},
	ConversionTypeWeight: {
		"lb":    {Name: "Pound", InSI: 0.453592},
		"kg":    {Name: "Kilogram", InSI: 1.0},
		"oz":    {Name: "Ounce", InSI: 0.0283495},
		"g":     {Name: "Gram", InSI: 0.001},
		"t":     {Name: "Ton", InSI: 1_000.0},
		"stone": {Name: "Stone", InSI: 6.35029},
	},
	ConversionTypeVolume: {
		"cup": {Name: "Cup", InSI: 0.24},
		"l":   {Name: "Liter", InSI: 1.0},
		"ml":  {Name: "Milliliter", InSI: 0.001},
		"gal": {Name: "Gallon", InSI: 3.78541},
		"pt":  {Name: "Pint", InSI: 0.473176},
		"qt":  {Name: "Quart", InSI: 0.946353},
	},
	ConversionTypeTime: {
		"s":     {Name: "Second", InSI: 1.0},
		"m":     {Name: "Minute", InSI: 60.0},
		"h":     {Name: "Hour", InSI: 3_600.0},
		"day":   {Name: "Day", InSI: 86_400.0},
		"week":  {Name: "Week", InSI: 604_800.0},
		"month": {Name: "Month", InSI: 2_629_800.0},
		"year":  {Name: "Year", InSI: 31_556_952.0},
	},
}

func Convert(value float64, fromUnit string, toUnit string) (float64, error) {
	var fromConversion, toConversion Conversion
	var found bool
	var isTemperature bool

	for conversionType, units := range unitMappings {
		if from, ok := units[fromUnit]; ok {
			if to, ok := units[toUnit]; ok {
				fromConversion = from
				toConversion = to
				found = true
				isTemperature = conversionType == ConversionTypeTemperature
				break
			}
		}
	}

	if !found {
		return 0, fmt.Errorf("units %s and/or %s not found or incompatible", fromUnit, toUnit)
	}

	if isTemperature {
		return convertTemperature(value, fromUnit, toUnit)
	}

	result := value * fromConversion.InSI / toConversion.InSI
	return result, nil
}

func convertTemperature(value float64, fromUnit string, toUnit string) (float64, error) {
	switch fromUnit {
	case "c":
		switch toUnit {
		case "c":
			return value, nil
		case "f":
			return (value * 9 / 5) + 32, nil
		case "k":
			return value + 273.15, nil
		}
	case "f":
		switch toUnit {
		case "f":
			return value, nil
		case "c":
			return (value - 32) * 5 / 9, nil
		case "k":
			return (value-32)*5/9 + 273.15, nil
		}
	case "k":
		switch toUnit {
		case "k":
			return value, nil
		case "c":
			return value - 273.15, nil
		case "f":
			return (value-273.15)*9/5 + 32, nil
		}
	}

	return 0, fmt.Errorf("invalid temperature conversion from %s to %s", fromUnit, toUnit)
}
