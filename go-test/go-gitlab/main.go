package main

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
	"log"
	"path"
)

func main() {
	url := "https://github.com/abc/dev-zt-api.2345.cn.git"

	fmt.Println(path.Base(url))
	split, file := path.Split(url)
	fmt.Println(path.Base(split))
	fmt.Println(file)
	fmt.Println()
	return

	git, err := gitlab.NewBasicAuthClient(
		"fanfei93",
		"19931206Ff!",
		gitlab.WithBaseURL("https://gitlab.com"),
	)
	if err != nil {
		log.Fatal(err)
	}

	//namespace := "piwik"
	//name := "piwik"
	//path := "https://gitlab.com/piwik/piwik"

	//project, _, err := git.Projects.ForkProject(1041501, &gitlab.ForkProjectOptions{
	//	//Namespace: gitlab.String(namespace),
	//	//Name:      gitlab.String(name),
	//	//Path:      gitlab.String(path),
	//})
	//if err != nil {
	//	panic("Projects.ForkProject returned error:" + err.Error())
	//}
	//
	//fmt.Println(project.Name)

	//获取项目
	//project, _, err := git.Projects.GetProject("justibabe/h", nil)
	//if err != nil {
	//	panic("获取项目失败："+err.Error())
	//}
	//fmt.Println(project.Name)
	//return

	// 创建分支
	_, _, err = git.Branches.CreateBranch("fanfei93/test", &gitlab.CreateBranchOptions{
		Branch: gitlab.String("fanfei-test1"),
		Ref:    gitlab.String("master"),
	})
	if err != nil {
		panic("create branch error:"+err.Error())
	}
	fmt.Println("create branch success")
	//fmt.Println(11111)
	//return

	//创建commit
	actions := make([]*gitlab.CommitAction,1)
	actions[0] = &gitlab.CommitAction{
		Action:          "update",
		FilePath:        "/test.md",
		PreviousPath:    "",
		Content:         "test11111",
		Encoding:        "",
		LastCommitID:    "",
		ExecuteFilemode: false,
	}
	commit, _, err := git.Commits.CreateCommit("fanfei93/test", &gitlab.CreateCommitOptions{
		Branch:        gitlab.String("fanfei-test1"),
		CommitMessage: gitlab.String("commit message"),
		StartBranch:   nil,
		StartSHA:      nil,
		StartProject:  nil,
		Actions:       actions,
		AuthorEmail:   nil,
		AuthorName:    nil,
		Stats:         nil,
		Force:         nil,
	})
	if err != nil {
		panic("commit file error: "+ err.Error())
	}
	fmt.Println(commit.ID)
	fmt.Println("create commit success")
	//return

	// 创建merge-request
	request, _, err := git.MergeRequests.CreateMergeRequest("fanfei93/test", &gitlab.CreateMergeRequestOptions{
		Title:              gitlab.String("merge test"),
		Description:        nil,
		SourceBranch:       gitlab.String("fanfei-test1"),
		TargetBranch:       gitlab.String("test-tag"),
		Labels:             nil,
		AssigneeID:         nil,
		AssigneeIDs:        nil,
		TargetProjectID:    nil,
		MilestoneID:        nil,
		RemoveSourceBranch: gitlab.Bool(true),
		Squash:             nil,
		AllowCollaboration: nil,
	})
	if err != nil {
		panic("merge request error: " + err.Error())
	}
	//fmt.Println(request)
	fmt.Println(request.WebURL)
	fmt.Println("Create mr success")
	return

	// List all projects
	//projects, _, err := git.Projects.ListProjects(nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(projects[0].Name)
	//
	//
	//opt := &gitlab.ListProjectsOptions{Search: gitlab.String("test")}
	//projects, _, err = git.Projects.ListProjects(opt)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(projects[0].Name)
	//
	//users, _, err := git.Users.ListUsers(&gitlab.ListUsersOptions{Search:gitlab.String("fanfei")})
	//fmt.Println(users[0].Name)
	//git, err := gitlab.NewClient("SASDLKJWBESLFIESGSDF")
	//if err != nil {
	//	log.Fatalf("Failed to create client: %v", err)
	//}
	//
	//users, _, err := git.Users.ListUsers(&gitlab.ListUsersOptions{})
	//fmt.Println(users)
	//
	//opt := &gitlab.ListProjectsOptions{Search: gitlab.String("t")}
	//projects, _, err := git.Projects.ListProjects(opt)
	//fmt.Println(projects)
	//return
	//
	//// Create new project
	//p := &gitlab.CreateProjectOptions{
	//	Name:                 gitlab.String("My Project"),
	//	Description:          gitlab.String("Just a test project to play with"),
	//	MergeRequestsEnabled: gitlab.Bool(true),
	//	SnippetsEnabled:      gitlab.Bool(true),
	//	Visibility:           gitlab.Visibility(gitlab.PublicVisibility),
	//}
	//project, _, err := git.Projects.CreateProject(p)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Add a new snippet
	//s := &gitlab.CreateProjectSnippetOptions{
	//	Title:           gitlab.String("Dummy Snippet"),
	//	FileName:        gitlab.String("snippet.go"),
	//	Content:         gitlab.String("package main...."),
	//	Visibility:      gitlab.Visibility(gitlab.PublicVisibility),
	//}
	//_, _, err = git.ProjectSnippets.CreateSnippet(project.ID, s)
	//if err != nil {
	//	log.Fatal(err)
	//}
}

