package memoryadmission

func unavailable() Observation {
	return Observation{Status: ObservationUnavailable, Reason: "unavailable"}
}

func boundedObservation(source string, availableMiB, totalMiB int64) Observation {
	observation := Observation{Status: ObservationObserved, Source: source, AvailableMiB: availableMiB, TotalMiB: totalMiB}
	if ValidateObservation(observation) != nil {
		return unavailable()
	}
	return observation
}
