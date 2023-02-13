package util

import (
	"time"

	"gopkg.in/redsync.v1"
)

type LockOption = redsync.Option

// LockSetExpiry can be used to set the expiry of a mutex to the given value.
func LockSetExpiry(expiry time.Duration) LockOption {
	return redsync.SetExpiry(expiry)
}

// LockSetTries can be used to set the number of times lock acquire is attempted.
func LockSetTries(tries int) LockOption {
	return redsync.SetTries(tries)
}

// LockSetRetryDelay can be used to set the amount of time to wait between retries.
func LockSetRetryDelay(delay time.Duration) LockOption {
	return redsync.SetRetryDelay(delay)
}

// LockSetDriftFactor can be used to set the clock drift factor.
func LockSetDriftFactor(factor float64) LockOption {
	return redsync.SetDriftFactor(factor)
}
