package main

import (
	"encoding/json"
	"fmt"
	"time"
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
