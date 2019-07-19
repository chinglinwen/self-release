package project

import (
	"fmt"
	"testing"
)

func TestGenerateEnvs(t *testing.T) {
	v := Envs{
		{Key: "key", Value: "value"},
		{Key: "key", Value: "value", Secret: "secret"},
	}
	s, err := GenerateEnvs(v)
	if err != nil {
		t.Error("gen env err", err)
		return
	}
	fmt.Printf("%q\n%v", s, s)
}

func TestGenerateVolumes(t *testing.T) {
	v := Volumes{
		{Name: "name", NFSPath: "nfs-path-value", NFSServer: "nfs-ip"},
	}
	s, err := GenerateVolumes(v)
	if err != nil {
		t.Error("gen volumes err", err)
		return
	}
	fmt.Printf("%q\n%v", s, s)
}

func TestGenerateMounts(t *testing.T) {
	v := Mounts{
		{VolumeName: "volume-name", MountPath: "mount-path-value"},
	}
	s, err := GenerateMounts(v)
	if err != nil {
		t.Error("gen mounts err", err)
		return
	}
	fmt.Printf("%q\n%v", s, s)
}

func TestGenerateVolumeMounts(t *testing.T) {
	v := VolumeMounts{
		{Volume{NFSPath: "nfs-path-value", NFSServer: "nfs-ip"}, Mount{MountPath: "mount-path-value"}},
	}
	vs, ms, err := GenerateVolumeMountss(v)
	if err != nil {
		t.Error("gen volume and mounts err", err)
		return
	}
	fmt.Printf("vs:\n%vms:\n%v", vs, ms)
}
