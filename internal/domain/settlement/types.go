package settlement

const (
	TypeMajorCity = "MajorCity"
	TypeCity      = "City"
	TypeVillage   = "Village"
	TypeAbandoned = "Abandoned"
)

func Classify(population float64) string {
	switch {
	case population >= 50000:
		return TypeMajorCity
	case population >= 10000:
		return TypeCity
	case population >= 1000:
		return TypeVillage
	default:
		return TypeAbandoned
	}
}
