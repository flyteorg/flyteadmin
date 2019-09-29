package auth

import (
	"encoding/base64"
	"fmt"
	"github.com/gorilla/securecookie"
	"testing"
)

func TestBasdf(t *testing.T)  {

	out := securecookie.GenerateRandomKey(32)
	fmt.Println(base64.RawStdEncoding.EncodeToString(out))
}

