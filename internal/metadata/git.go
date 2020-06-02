package metadata

import (
	"fmt"

	util "github.com/applandinc/appland-cli/internal/util"
	jsonpatch "github.com/evanphx/json-patch"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

type GitMetadata struct {
	Repository string   `json:"repository,omitempty"`
	Commit     string   `json:"commit,omitempty"`
	Branch     string   `json:"branch,omitempty"`
	Status     []string `json:"status,omitempty"`
	Tag        string   `json:"annotated_tag,omitempty"`
}

var metadataCache = map[string]*GitMetadata{}

func GetGitMetadata(path string) (*GitMetadata, error) {
	repositoryInfo, err := util.GetRepository(path)
	if err != nil {
		return nil, err
	}

	existingMetadata := metadataCache[repositoryInfo.Path]
	if existingMetadata != nil {
		return existingMetadata, nil
	}

	gitMetadata, err := collectGitMetadata(repositoryInfo.Repository)
	if err != nil {
		return nil, err
	}

	metadataCache[repositoryInfo.Path] = gitMetadata

	return gitMetadata, nil
}

func (git *GitMetadata) AsPatch() (*jsonpatch.Patch, error) {
	patch, err := util.BuildPatch("replace", "/metadata/git", git)
	if err != nil {
		return nil, err
	}

	return &patch, nil
}

// findReference currently only resolves references to tags or commits
func findReference(repo *git.Repository, hash plumbing.Hash, iter storer.ReferenceIter) (*plumbing.Reference, error) {
	var refMatch *plumbing.Reference

	err := iter.ForEach(func(ref *plumbing.Reference) error {
		var commit *object.Commit
		obj, err := repo.TagObject(ref.Hash())
		switch err {
		case nil:
			commit, err = obj.Commit()
			if err != nil {
				return fmt.Errorf("failed to get commit from tag (tag %s): %w", ref.Hash(), err)
			}
		case plumbing.ErrObjectNotFound:
			commit, err = repo.CommitObject(ref.Hash())
			if err != nil {
				return fmt.Errorf("failed to get commit from ref (ref %s): %w", ref.Hash(), err)
			}
		default:
			return err
		}

		if commit.Hash.String() == hash.String() {
			refMatch = ref

			// returning an error here stops further iteration
			// with refMatch set, it will be nulled out later on
			return fmt.Errorf("")
		}

		return nil
	})

	// if we have a match we shouldn't have any errors
	if refMatch != nil {
		err = nil
	}

	if refMatch == nil && err == nil {
		err = fmt.Errorf("not found (%v)", hash.String())
	}

	return refMatch, err
}

func collectGitMetadata(repo *git.Repository) (*GitMetadata, error) {
	metadata := &GitMetadata{}
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve HEAD: %w", err)
	}
	metadata.Commit = head.Hash().String()

	headReference := head.Name()
	if headReference.IsBranch() {
		metadata.Branch = headReference.Short()
	} else {
		// handle detached HEAD
		commit, err := repo.CommitObject(head.Hash())
		if err != nil {
			return nil, fmt.Errorf("failed to read commit at HEAD")
		}

		// assume HEAD is a merge commit such as one created by GitHub
		// this attempts to retrieve a branch name from the last parent and as such,
		// this doesn't support multi-merge
		lastParentIndex := commit.NumParents() - 1
		if lastParentIndex == 1 {
			parentCommit, err := commit.Parent(lastParentIndex)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve parent commit for HEAD")
			}

			branches, err := repo.Branches()
			if err != nil {
				return nil, fmt.Errorf("failed to read branches from repository: %w", err)
			}

			ref, err := findReference(repo, parentCommit.Hash, branches)
			if err != nil {
				return nil, fmt.Errorf("failed to find reference: %w", err)
			}

			if ref.Name().IsBranch() {
				metadata.Branch = ref.Name().Short()

				// use the parent reference instead of HEAD because we're unlikely to
				// find tags on a detached HEAD
				head = ref
				fmt.Printf("merge commit in HEAD, assuming branch %s\n", metadata.Branch)
			}
		}
	}

	remote, err := repo.Remote("origin")
	if err == nil {
		remoteURLs := remote.Config().URLs
		if len(remoteURLs) > 0 {
			metadata.Repository = remoteURLs[0]
		}
	}

	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to read tags from repository: %w", err)
	}

	ref, err := findReference(repo, head.Hash(), tags)
	if err == nil {
		metadata.Tag = ref.Name().Short()
	}

	// The following comment block is functional code, but way too slow
	// I'm disabling this until there's a patch.
	// -DB
	/*
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, err
		}

		// this call is slow. very slow.
		status, err := worktree.Status()
		if err != nil {
			return nil, err
		}

		for filePath, fileStatus := range status {
			statusString := fmt.Sprintf("%c %s", fileStatus.Worktree, filePath)
			metadata.Status = append(metadata.Status, statusString)
		}
	*/

	return metadata, nil
}
