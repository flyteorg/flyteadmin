//build integration

// This is an integration test because the token will show up as expired, you will need a live token

package auth

//func TestVerify(t *testing.T) {
//	token, err := jwt.Parse(token, GetKey)
//	assert.NoError(t, err)
//	if err != nil {
//		panic(err)
//	}
//	claims := token.Claims.(jwt.MapClaims)
//	for key, value := range claims {
//		fmt.Printf("%s\t%v\n", key, value)
//	}
//}
