package classify

// Classify returns the danger Level of the given shell command string.
// Classification is purely pattern-based (no AI call) for low latency.
func Classify(cmd string) Level {
	for _, p := range dangerPatterns {
		if p.MatchString(cmd) {
			return Danger
		}
	}
	for _, p := range cautionPatterns {
		if p.MatchString(cmd) {
			return Caution
		}
	}
	return Safe
}
