package utils

func IsValidInterval(interval string) bool {
	switch interval {
	case "Minute", "Hour", "Day", "Week", "Month", "Quarter", "Year":
		return true
	default:
		return false
	}
}

