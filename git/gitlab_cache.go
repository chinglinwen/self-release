package git

import gitlab "github.com/xanzy/go-gitlab"

// since projects doesn't change frequent, let's reduce fetch time
var projectsCache map[string][]*gitlab.Project
