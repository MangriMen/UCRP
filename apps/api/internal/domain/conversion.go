package domain

import "math"

func shareOf(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func (c Conversion) Rate() float64 {
	return shareOf(c.Conversions, c.Recipients)
}

func (c Conversion) ControlRate() float64 {
	return shareOf(c.ControlConversions, c.ControlGroup)
}

func (c Conversion) Lift() float64 {
	if c.ControlGroup == 0 {
		return 0
	}
	return c.Rate() - c.ControlRate()
}

func (c Conversion) PValue() (float64, bool) {
	if c.Recipients == 0 || c.ControlGroup == 0 {
		return 0, false
	}
	pooled := shareOf(c.Conversions+c.ControlConversions, c.Recipients+c.ControlGroup)
	standardError := math.Sqrt(pooled * (1 - pooled) * (1/float64(c.Recipients) + 1/float64(c.ControlGroup)))
	if standardError == 0 {
		return 1, true
	}
	z := math.Abs(c.Lift()) / standardError
	return math.Erfc(z / math.Sqrt2), true
}
