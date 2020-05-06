package metadata

import (
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectGitMetadata(t *testing.T) {
	fs := memfs.New()
	repo, err := git.Init(memory.NewStorage(), fs)
	require.Nil(t, err)

	originURL := "https://myorg.com/example.git"
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{originURL},
	})

	err = util.WriteFile(fs, "foo", []byte("foo"), 0755)
	require.Nil(t, err)

	w, err := repo.Worktree()
	require.Nil(t, err)

	_, err = w.Add("foo")
	require.Nil(t, err)

	_, err = w.Commit("foo", &git.CommitOptions{Author: &object.Signature{
		Name:  "foo",
		Email: "foo@foo.foo",
		When:  time.Now(),
	}})
	require.Nil(t, err)

	head, err := repo.Head()
	require.Nil(t, err)

	tagName := "v0.0.0"
	tagRef, err := repo.CreateTag(tagName, head.Hash(), nil)
	require.Nil(t, err)
	require.NotNil(t, tagRef)

	metadata, err := collectGitMetadata(repo)
	require.Nil(t, err)
	require.NotNil(t, metadata)

	assert := assert.New(t)
	assert.Equal(head.Hash().String(), metadata.Commit)
	assert.Equal("master", metadata.Branch)
	assert.Equal(originURL, metadata.Repository)
	assert.Equal(tagName, metadata.Tag)
}
