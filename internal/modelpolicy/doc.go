// Package modelpolicy selects an explicitly configured role profile from
// observable task signals. It is deliberately pure: it neither discovers
// model availability nor sends provider work. A caller must admit the returned
// exact profile through its normal runtime and effect controls.
//
// The package has no value-of-information or JEV stopping policy. There is no
// live evidence yet that either improves task quality or reduces intervention;
// adding one before that measurement would turn a hypothesis into a routing
// rule. Fixture baseline cases exercise only deterministic policy behavior and
// make no quality, cost, or latency claim.
package modelpolicy
