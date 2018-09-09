package pam

//#include "pam.h"
//#cgo LDFLAGS: -L. -lpam -lpam_misc
import "C"

// Authenticate checks a user, password combination
func Authenticate(user, password string) bool {
	ret := C.check_password(C.CString(user), C.CString(password))
	return ret == 1
}

// AuthenticateCurrentUser checks the specified password
func AuthenticateCurrentUser(password string) bool {
	ret := C.check_current_user(C.CString(password))
	return ret == 1
}
