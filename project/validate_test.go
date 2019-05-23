package project

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestValidateByKubeval(t *testing.T) {
	r, err := ValidateByKubeval(examplefs, "fs")
	if err != nil {
		t.Error("validate err", err)
		return
	}
	spew.Dump("r", r)
}

func TestValidateByKubectl(t *testing.T) {
	out, err := ValidateByKubectl(examplefs, "fs")
	if err != nil {
		t.Errorf("validate err: %v\n", err)
		return
	}
	fmt.Println("out:", out)
}

var examplefs = `
# Service
apiVersion: v2
kind: Service
metadata:
  name: fs
  namespace: yunwei
spec:
  ports:
  - name: web
    targetPort: 8000
    port: 80
  selector:
    app: fs
  sessionAffinity: ClientIP

# Ingress
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: fs
  namespace: yunwei
  annotations:
    traefik.ingress.kubernetes.io/frontend-entry-points: http,https
spec:
  rules:
  - host: fs.devops.haodai.net
    http:
      paths:
      - path: /
        backend:
          serviceName: fs
          servicePort: web

# Deployment
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: fs
  namespace: yunwei
  labels:
    app: fs
spec:
  replicas: 1
  selector:
    matchLabels:
      app: fs
  template:
    metadata:
      labels:
        app: fs
    spec:
      nodeSelector:
        func: "monitor"
      containers:
        - name: fs
          image: chinglinwen/fs
          ports:
          - containerPort: 8000
          volumeMounts:
          - mountPath: /data
            name: data
      volumes:
        - name: data
          hostPath:
            path: /data/k8s/fs
            type: Directory
`
