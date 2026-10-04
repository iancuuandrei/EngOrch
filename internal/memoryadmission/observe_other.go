//go:build !windows && !linux

package memoryadmission

func observe() Observation { return unavailable() }
