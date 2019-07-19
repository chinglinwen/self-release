package project

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/chinglinwen/log"
)

/*
envs
        ## mysql
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: local-3306-database-username
              key: host
        - name: DB_PORT
          valueFrom:
            secretKeyRef:
              name: local-3306-database-username
              key: port

volumes
      - name: fanqiyu-uploads
        nfs:
          path: /data/staticfile_yjr/fanqiyu
          server: 172.31.83.26
mounts
        - name: fanqiyu-uploads
          mountPath: /apps/fanqiyu/public/uploads
*/

const (
	EnvTmpl = `
{{range .}}
        - name: {{ .Key }}
{{if .Secret}}          valueFrom:
            secretKeyRef:
              name: {{ .Secret }}
              key: {{ .Value }}
{{else}}          value: "{{ .Value }}"{{end}}{{end}}
`
	VolumeTmpl = `
{{range .}}
      - name: {{ .Name }}
        nfs:
          path: {{ .NFSPath }}
          server: {{ .NFSServer }}
{{end}}
`

	VolumeMountTmpl = `
{{range .}}
        - name: {{ .VolumeName }}
          mountPath: {{ .MountPath }}
{{end}}
`
)

type Env struct {
	Key    string
	Value  string
	Secret string
}

type Envs []Env

type Volume struct {
	Name      string
	NFSPath   string
	NFSServer string
}

type Volumes []Volume

type Mount struct {
	VolumeName string
	MountPath  string
}

type Mounts []Mount

type VolumeMount struct {
	Volume
	Mount
}

type VolumeMounts []VolumeMount

var (
	envTmpl         *template.Template
	volumeTmpl      *template.Template
	volumeMountTmpl *template.Template
)

func init() {
	var err error
	envTmpl, err = template.New("envs").Parse(EnvTmpl)
	if err != nil {
		log.Fatal(err)
	}
	volumeTmpl, err = template.New("volumes").Parse(VolumeTmpl)
	if err != nil {
		log.Fatal(err)
	}
	volumeMountTmpl, err = template.New("volumeMounts").Parse(VolumeMountTmpl)
	if err != nil {
		log.Fatal(err)
	}
}
func GenerateEnvs(v Envs) (s string, err error) {
	if v == nil {
		return
	}
	buf := &bytes.Buffer{}
	err = envTmpl.Execute(buf, v)
	s = buf.String()
	return
}

func GenerateVolumes(v Volumes) (s string, err error) {
	if v == nil {
		return
	}
	buf := &bytes.Buffer{}
	err = volumeTmpl.Execute(buf, v)
	s = buf.String()
	return
}

func GenerateMounts(v Mounts) (s string, err error) {
	if v == nil {
		return
	}
	buf := &bytes.Buffer{}
	err = volumeMountTmpl.Execute(buf, v)
	s = buf.String()
	return
}

func GenerateVolumeMountss(vms VolumeMounts) (vs, ms string, err error) {
	if vms == nil {
		return
	}
	volumes, mounts := Volumes{}, Mounts{}
	for _, v := range vms {
		if v.MountPath == "" {
			continue
		}
		if v.Name != "" && v.VolumeName != "" {
			v.Volume.Name = v.Mount.MountPath
		}
		volumes = append(volumes, v.Volume)
		mounts = append(mounts, v.Mount)
	}

	vs, err = GenerateVolumes(volumes)
	if err != nil {
		err = fmt.Errorf("generate volumes err: %v", err)
		return
	}
	ms, err = GenerateMounts(mounts)
	if err != nil {
		err = fmt.Errorf("generate mounts err: %v", err)
		return
	}
	return
}
