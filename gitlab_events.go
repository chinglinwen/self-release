package main

import (
	"encoding/json"
	"fmt"
	"time"
	"wen/self-release/pkg/sse"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
)

func ParseEvent(eventName string, payload []byte) (data interface{}, err error) {
	switch eventName {
	case "push":
		data = &PushEvent{}
	case "tag_push":
		data = &TagPushEvent{}
	case "repository_update":
		data = &RepositoryUpdateEvent{}
	default:
		err = fmt.Errorf("unknown event type: %v", eventName)
		return
	}

	err = json.Unmarshal(payload, &data)
	return
}

const TimeLayout = "2006-1-2_15:04:05"
const IDTimeLayout = "20060102.150405"

func (event *PushEvent) GetInfo() (e *sse.EventInfo, err error) {
	e = &sse.EventInfo{}
	e.Project = event.Project.PathWithNamespace
	e.Branch = parseBranch(event.Ref)

	if e.Branch == errParseRefs {
		err = fmt.Errorf("project: %v, parse branch err for refs: %v", e.Project, event.Ref)
		return
	}
	// e.Env = projectpkg.GetEnvFromBranch(e.Branch)
	e.UserName = event.UserUsername
	e.UserEmail = event.UserEmail

	// use time and commit-id together
	e.CommitID = time.Now().Format(IDTimeLayout)
	n := len(event.Commits)
	if n > 0 {
		if len(event.Commits[n-1].ID) >= 8 {
			e.CommitID += ".id:" + event.Commits[n-1].ID[:8]
		}
		e.Message = fmt.Sprintf("[gitlab tag] %v", event.Commits[0].Message)
	}

	e.Time = time.Now().Format(TimeLayout)

	return
}

func (event *TagPushEvent) GetInfo() (e *sse.EventInfo, err error) {
	e = &sse.EventInfo{}
	e.Project = event.Project.PathWithNamespace
	e.Branch = parseBranch(event.Ref)

	if e.Branch == errParseRefs {
		err = fmt.Errorf("project: %v, parse branch err for refs: %v", e.Project, event.Ref)
		return
	}
	// e.Env = projectpkg.GetEnvFromBranch(branch) ?
	e.UserName = event.UserUsername
	e.UserEmail = event.UserEmail

	// use time and commit-id together
	e.CommitID = time.Now().Format(IDTimeLayout)
	n := len(event.Commits)
	if n > 0 {
		if len(event.Commits[n-1].ID) >= 8 {
			e.CommitID += ".id:" + event.Commits[n-1].ID[:8]
		}
	}

	e.Message = fmt.Sprintf("[gitlab tag] %v", event.Message) // release message
	e.Time = time.Now().Format(TimeLayout)

	return
}

type Eventer interface {
	GetInfo() (e *sse.EventInfo, err error)
}

func GetEventInfo(event Eventer) (e *sse.EventInfo, err error) {
	return event.GetInfo()
}

func GetEventInfoToMap(event Eventer) (autoenv map[string]string, err error) {
	e, err := event.GetInfo()
	if err != nil {
		return
	}
	return EventInfoToMap(e)
}

// name need to be different for different env
// makes parse project name harder?
var projectYamlTmpl = `
apiVersion: project.haodai.com/v1alpha1
kind: Project
metadata:
  name: %v-%v
  namespace: %v
spec:
  version: "%v"
  userName: "%v"
  userEmail: "%v"
  releaseMessage: "%v"
  releaseAt: "%v"
`

// must provide enough info for EventInfoToMap later
func EventInfoToProjectYaml(e *sse.EventInfo) (body string, err error) {
	ns, name, err := projectpkg.GetProjectName(e.Project)
	if err != nil {
		err = fmt.Errorf("parse project name for %q, err: %v", e.Project, err)
		return
	}
	env := projectpkg.GetEnvFromBranchOrCommitID(e.Project, e.Branch)
	version := e.Branch

	// for test env, change version to commitid if from gitlab event
	if env == projectpkg.TEST && e.CommitID != "" {
		// so test image changed ( otherwise always the same )
		version = e.CommitID
	}

	if e.Time == "" {
		e.Time = time.Now().Format(TimeLayout)
	}
	log.Printf("construct yaml: project: %v, env: %v, version: %v\n", e.Project, env, version)

	// convert info to version?
	body = fmt.Sprintf(projectYamlTmpl, name, env, ns, version,
		e.UserName, e.UserEmail, e.Message, e.Time)
	return
}

func EventInfoToMap(e *sse.EventInfo) (autoenv map[string]string, err error) {

	namespace, projectName, err := projectpkg.GetProjectName(e.Project)
	if err != nil {
		err = fmt.Errorf("parse project name for %q, err: %v", e.Project, err)
		return
	}

	// is this needed, we often don't need overwrite env by manual?
	if e.Env == "" {
		e.Env = projectpkg.GetEnvFromBranchOrCommitID(e.Project, e.Branch)
	}
	if e.Time == "" {
		e.Time = time.Now().Format(TimeLayout)
	}

	autoenv = make(map[string]string)
	autoenv["CI_PROJECT_PATH"] = e.Project
	// autoenv["CI_BRANCH"] = e.Branch // don't need this anymore
	autoenv["CI_ENV"] = e.Env
	autoenv["CI_NAMESPACE"] = namespace
	autoenv["CI_PROJECT_NAME"] = projectName
	autoenv["CI_PROJECT_NAME_WITH_ENV"] = projectName + "-" + e.Env
	autoenv["CI_REPLICAS"] = "1" // config.env has higher priority to overwrite this

	// calc by version? tag or commitid
	autoenv["CI_IMAGE"] = projectpkg.GetImage(e.Project, e.Branch) // or using project_path

	autoenv["CI_USER_NAME"] = e.UserName
	autoenv["CI_USER_EMAIL"] = e.UserEmail
	autoenv["CI_MSG"] = e.Message
	autoenv["CI_TIME"] = e.Time

	return
}

type PushEvent struct {
	After       string `json:"after"`
	Before      string `json:"before"`
	CheckoutSha string `json:"checkout_sha"`
	Commits     []struct {
		Added  []string `json:"added"`
		Author struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"author"`
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Modified  []string  `json:"modified"`
		Removed   []string  `json:"removed"`
		Timestamp time.Time `json:"timestamp"`
		URL       string    `json:"url"`
	} `json:"commits"`
	EventName  string `json:"event_name"`
	Message    string `json:"message"`
	ObjectKind string `json:"object_kind"`
	Project    struct {
		AvatarURL         string `json:"avatar_url"`
		CiConfigPath      string `json:"ci_config_path"`
		DefaultBranch     string `json:"default_branch"`
		Description       string `json:"description"`
		GitHTTPURL        string `json:"git_http_url"`
		GitSSHURL         string `json:"git_ssh_url"`
		Homepage          string `json:"homepage"`
		HTTPURL           string `json:"http_url"`
		ID                int    `json:"id"`
		Name              string `json:"name"`
		Namespace         string `json:"namespace"`
		PathWithNamespace string `json:"path_with_namespace"`
		SSHURL            string `json:"ssh_url"`
		URL               string `json:"url"`
		VisibilityLevel   int    `json:"visibility_level"`
		WebURL            string `json:"web_url"`
	} `json:"project"`
	ProjectID  int    `json:"project_id"`
	Ref        string `json:"ref"`
	Repository struct {
		Description     string `json:"description"`
		GitHTTPURL      string `json:"git_http_url"`
		GitSSHURL       string `json:"git_ssh_url"`
		Homepage        string `json:"homepage"`
		Name            string `json:"name"`
		URL             string `json:"url"`
		VisibilityLevel int    `json:"visibility_level"`
	} `json:"repository"`
	TotalCommitsCount int    `json:"total_commits_count"`
	UserAvatar        string `json:"user_avatar"`
	UserEmail         string `json:"user_email"`
	UserID            int    `json:"user_id"`
	UserName          string `json:"user_name"`
	UserUsername      string `json:"user_username"`
}

type TagPushEvent struct {
	ObjectKind   string `json:"object_kind"`
	EventName    string `json:"event_name"`
	Before       string `json:"before"`
	After        string `json:"after"`
	Ref          string `json:"ref"`
	CheckoutSha  string `json:"checkout_sha"`
	Message      string `json:"message"`
	UserID       int    `json:"user_id"`
	UserName     string `json:"user_name"`
	UserUsername string `json:"user_username"`
	UserEmail    string `json:"user_email"`
	UserAvatar   string `json:"user_avatar"`
	ProjectID    int    `json:"project_id"`
	Project      struct {
		ID                int    `json:"id"`
		Name              string `json:"name"`
		Description       string `json:"description"`
		WebURL            string `json:"web_url"`
		AvatarURL         string `json:"avatar_url"`
		GitSSHURL         string `json:"git_ssh_url"`
		GitHTTPURL        string `json:"git_http_url"`
		Namespace         string `json:"namespace"`
		VisibilityLevel   int    `json:"visibility_level"`
		PathWithNamespace string `json:"path_with_namespace"`
		DefaultBranch     string `json:"default_branch"`
		CiConfigPath      string `json:"ci_config_path"`
		Homepage          string `json:"homepage"`
		URL               string `json:"url"`
		SSHURL            string `json:"ssh_url"`
		HTTPURL           string `json:"http_url"`
	} `json:"project"`
	Commits []struct {
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		URL       string    `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Added    []string `json:"added"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
	} `json:"commits"`
	TotalCommitsCount int `json:"total_commits_count"`
	Repository        struct {
		Name            string `json:"name"`
		URL             string `json:"url"`
		Description     string `json:"description"`
		Homepage        string `json:"homepage"`
		GitHTTPURL      string `json:"git_http_url"`
		GitSSHURL       string `json:"git_ssh_url"`
		VisibilityLevel int    `json:"visibility_level"`
	} `json:"repository"`
}

type RepositoryUpdateEvent struct {
	Changes []struct {
		After  string `json:"after"`
		Before string `json:"before"`
		Ref    string `json:"ref"`
	} `json:"changes"`
	EventName string `json:"event_name"`
	Project   struct {
		AvatarURL         string `json:"avatar_url"`
		CiConfigPath      string `json:"ci_config_path"`
		DefaultBranch     string `json:"default_branch"`
		Description       string `json:"description"`
		GitHTTPURL        string `json:"git_http_url"`
		GitSSHURL         string `json:"git_ssh_url"`
		Homepage          string `json:"homepage"`
		HTTPURL           string `json:"http_url"`
		ID                int    `json:"id"`
		Name              string `json:"name"`
		Namespace         string `json:"namespace"`
		PathWithNamespace string `json:"path_with_namespace"`
		SSHURL            string `json:"ssh_url"`
		URL               string `json:"url"`
		VisibilityLevel   int    `json:"visibility_level"`
		WebURL            string `json:"web_url"`
	} `json:"project"`
	ProjectID  int      `json:"project_id"`
	Refs       []string `json:"refs"`
	UserAvatar string   `json:"user_avatar"`
	UserEmail  string   `json:"user_email"`
	UserID     int      `json:"user_id"`
	UserName   string   `json:"user_name"`
}
