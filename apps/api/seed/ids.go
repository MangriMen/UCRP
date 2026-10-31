package seed

import "github.com/google/uuid"

var demoNamespace = uuid.MustParse("b1f0f2a4-6c1d-5e8a-9f3b-2d7c4a1e6b90")

func demoID(kind, name string) string {
	return uuid.NewSHA1(demoNamespace, []byte(kind+"/"+name)).String()
}
