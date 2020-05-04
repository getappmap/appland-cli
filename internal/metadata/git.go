package metadata

import (
	"fmt"

	util "github.com/applandinc/appland-cli/internal/util"
	jsonpatch "github.com/evanphx/json-patch"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type GitMetadata struct {
	Repository string   `json:"repository,omitempty"`
	Commit     string   `json:"commit,omitempty"`
	Branch     string   `json:"branch,omitempty"`
	Status     []string `json:"status,omitempty"`
	Tag        string   `json:"annotated_tag,omitempty"`
}

var metadataCache = map[string]*jsonpatch.Patch{}

func GetGitMetadata(path string) (*jsonpatch.Patch, error) {
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

	patch, err := util.BuildPatch("replace", "/metadata/git", gitMetadata)
	if err != nil {
		return nil, err
	}

	metadataCache[repositoryInfo.Path] = &patch

	return &patch, nil
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
		metadata.Branch = head.Name().Short()
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

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		var commit *object.Commit

		if ref.Target().IsTag() {
			tag, err := repo.TagObject(ref.Hash())
			if err != nil {
				return fmt.Errorf("failed to resolve annotated tag (ref %s): %w", ref.Hash(), err)
			}

			commit, err = tag.Commit()
			if err != nil {
				return fmt.Errorf("failed to get commit from annotated tag (ref %s): %w", ref.Hash(), err)
			}
		} else {
			tagObj, err := repo.Object(plumbing.TagObject, ref.Hash())
			if err != nil {
				return fmt.Errorf("failed to resolve lightweight tag object (ref %s): %w", ref.Hash(), err)
			}

			if tag, ok := tagObj.(*object.Tag); ok {
				commit, err = tag.Commit()
				if err != nil {
					return fmt.Errorf("failed to get commit from lightweight tag (ref %s): %w", ref.Hash(), err)
				}
			}
		}

		if commit.Hash.String() == head.Hash().String() {
			metadata.Tag = ref.Name().Short()
		}

		return nil
	})
	if err != nil {
		return nil, err
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
