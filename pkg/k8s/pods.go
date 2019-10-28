package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	prettyjson "github.com/hokaccha/go-prettyjson"
)

type PodInfo struct {
	Name      string `json:"name,omitempty"`
	PodName   string `json:"pod_name,omitempty"`
	Env       string `json:"env,omitempty"`
	GitName   string `json:"git_name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	IP        string `json:"ip,omitempty"`
	Node      string `json:"node,omitempty"`
	Phase     string `json:"phase,omitempty"`
	Starttime string `json:"starttime,omitempty"`
	Message   string `json:"message,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

const TimeLayout = "2006-1-2_15:04:05"

func PodListInfo(project string) (pods []PodInfo, err error) {
	k8sgit := strings.Replace(project, "_", "-", -1)

	ns, _ := getnsrepo(project)
	ps, err := podList(ns)
	if err != nil {
		return
	}
	for _, v := range ps {

		podname := v.Metadata.GetName()
		ns := v.Metadata.GetNamespace()
		name, env := getNameEnv(podname)

		st := v.GetStatus()
		phase := st.GetPhase()
		node := st.GetHostIP()
		ip := st.GetPodIP()

		starttime := time.Unix(st.GetStartTime().GetSeconds(),
			int64(st.GetStartTime().GetNanos())).Local().Format(TimeLayout)
		//
		// pretty("status")

		s := fmt.Sprintf("%v/%v", ns, podname)
		if hasPrefix(s, k8sgit) {
			pods = append(pods, PodInfo{
				Name:      name,
				PodName:   podname,
				Env:       env,
				GitName:   getGitName(ns, name),
				Namespace: ns,
				Node:      node,
				IP:        ip,
				Phase:     phase,
				Starttime: starttime,
				Message:   st.GetMessage(),
				Reason:    st.GetReason(),
			})
		}
	}
	if len(pods) == 0 {
		err = fmt.Errorf("no pods found")
	}
	return
}

func pretty(prefix string, a interface{}) {
	out, _ := prettyjson.Marshal(a)
	fmt.Printf("%v: %s\n", prefix, out)
}

func hasPrefix(s, k8sgit string) bool {
	if strings.HasPrefix(s, k8sgit+"-"+ONLINE) {
		return true
	}
	if strings.HasPrefix(s, k8sgit+"-"+PRE) {
		return true
	}
	if strings.HasPrefix(s, k8sgit+"-"+TEST) {
		return true
	}
	return false
}

const (
	ONLINE = "online"
	PRE    = "pre"
	TEST   = "test"
)

func getNameEnv(podname string) (name, env string) {
	s := strings.Split(podname, "-")
	n := len(s)
	if n > 2 {
		name = strings.Join(s[:n-2], "-")
	}
	var e string
	if n > 3 {
		e = s[n-3]
	}

	switch e {
	case ONLINE:
		env = ONLINE
	case PRE:
		env = PRE
	case TEST:
		env = TEST
	default:
	}
	return
}

func getGitName(ns, name string) string {
	return ns + "/" + trimEnv(name)
}

func trimEnv(name string) string {
	name = strings.TrimSuffix(name, "-"+ONLINE)
	name = strings.TrimSuffix(name, "-"+PRE)
	name = strings.TrimSuffix(name, "-"+TEST)
	return name
}

func podList(ns string) (pods []*corev1.Pod, err error) {
	var slist corev1.PodList
	err = client.List(context.Background(), ns, &slist)
	if err != nil {
		err = fmt.Errorf("get pods err %v", err)
		return
	}
	pods = slist.GetItems()
	return
}

func getnsrepo(git string) (ns, repo string) {
	k8sgit := strings.Replace(git, "_", "-", -1)
	giturl := strings.Split(k8sgit, "/")
	if len(giturl) >= 1 {
		ns = giturl[0]
	}
	if len(giturl) >= 2 {
		repo = giturl[1]
	}
	return
}
