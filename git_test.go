package cali

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContainerName(t *testing.T) {
	t.Log("Testing simple example")
	name, err := GitCheckoutConfig{
		Repo:    "repo",
		Branch:  "branch",
		RelPath: ".",
	}.GetContainerName()
	if assert.NoError(t, err) {
		assert.Equal(t, "data_repo_._branch_89bb6900736548ebd6455d0ab07aa5fe", name)
	}

	t.Log("Testing branches and paths with slashes")
	name, err = GitCheckoutConfig{
		Repo:    "repo",
		Branch:  "branch/with-slash",
		RelPath: "path/containing/slashes",
	}.GetContainerName()
	if assert.NoError(t, err) {
		assert.Equal(t, "data_repo_path-containing-slashes_branch-with-slash_89bb6900736548ebd6455d0ab07aa5fe", name)
	}
}

func TestRepoNameFromUrl(t *testing.T) {
	t.Log("Testing an example git url")
	name, err := repoNameFromUrl("git@github.com:skybet/cali.git")
	if assert.NoError(t, err) {
		assert.Equal(t, "github-com-skybet-cali", name)
	}

	t.Log("Testing an example https url")
	name, err = repoNameFromUrl("https://github.com/skybet/cali.git")
	if assert.NoError(t, err) {
		assert.Equal(t, "github-com-skybet-cali", name)
	}

	t.Log("Testing an example git url with ssh protocol")
	name, err = repoNameFromUrl("ssh://git@github.com/skybet/cali.git")
	if assert.NoError(t, err) {
		assert.Equal(t, "github-com-skybet-cali", name)
	}
}
