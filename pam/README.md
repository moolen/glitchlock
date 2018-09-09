# PAM

Check a user/password combination using PAM.

You need to install `<security/pam_appl.h>` and `<security/pam_misc.h>`.

## Example

```go
package main

import (
	"github.com/moolen/glitchlock/pam"
)

func main() {
	if pam.AuthenticateCurrentUser("1234") {
        fmt.Println("password matches")
    }else{
        fmt.Println("wrong password")
    }
    if pam.Authenticate("my-user", "1234") {
        fmt.Println("password matches")
    }else{
        fmt.Println("wrong password")
    }
}


```
