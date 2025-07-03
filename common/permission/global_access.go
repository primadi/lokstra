package permission

var globalAccessLocked = false

func LockGlobalAccess() {
	globalAccessLocked = true
}

func GlobalAccessLocked() bool {
	return globalAccessLocked
}
