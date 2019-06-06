package project

import (
	"fmt"
	"testing"
)

// go test -timeout 60s wen/self-release/project -run TestBuild -v -count=1
func TestCheckOrCreateNamespace(t *testing.T) {
	out, err := CheckOrCreateNamespace("t")
	if err != nil {
		t.Error("CheckOrCreateNamespace", err)
		return
	}
	fmt.Println(out)
}

/*

kubectl apply -f - <<eof
---
apiVersion: v1
kind: Namespace
metadata:
  name: t
---
# harborkey
apiVersion: v1
data:
  .dockerconfigjson: eyJhdXRocyI6eyJoYXJib3IuaGFvZGFpLm5ldCI6eyJ1c2VybmFtZSI6ImRldnVzZXIiLCJwYXNzd29yZCI6IkxuMjhvaHlEbiIsImVtYWlsIjoieXVud2VpQGhhb2RhaS5uZXQiLCJhdXRoIjoiWkdWMmRYTmxjanBNYmpJNGIyaDVSRzQ9In19fQ==
kind: Secret
metadata:
  name: devuser-harborkey
  namespace: t
type: kubernetes.io/dockerconfigjson
eof
*/
