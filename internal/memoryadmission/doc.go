// Package memoryadmission observes host memory and derives conservative worker
// ceilings for future admission. Observations are not worker usage measurements;
// unavailable observations retain uncertainty, and decisions never revoke an
// already admitted effect. Controllers persist decisions for exact replay.
package memoryadmission
