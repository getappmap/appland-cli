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

type gitMetadataBuilder struct {
	repository *git.Repository
	commit     *plumbing.Reference
	branch     plumbing.ReferenceName
	tag        plumbing.ReferenceName
}

// FindReference currently only resolves references to tags or commits
func (gm *gitMetadataBuilder) FindReference(hash plumbing.Hash, iter storer.ReferenceIter) (*plumbing.Reference, error) {
	var refMatch *plumbing.Reference

	err := iter.ForEach(func(ref *plumbing.Reference) error {
		var commit *object.Commit
		obj, err := gm.repository.TagObject(ref.Hash())
		switch err {
		case nil:
			commit, err = obj.Commit()
			if err != nil {
				return fmt.Errorf("failed to get commit from tag (tag %s): %w", ref.Hash(), err)
			}
		case plumbing.ErrObjectNotFound:
			commit, err = gm.repository.CommitObject(ref.Hash())
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

func (gm *gitMetadataBuilder) BranchName() string {
	return gm.branch.Short()
}

func (gm *gitMetadataBuilder) CommitHash() string {
	if gm.commit == nil {
		return ""
	}

	return gm.commit.Hash().String()
}

func (gm *gitMetadataBuilder) TagName() string {
	return gm.tag.Short()
}

func (gm *gitMetadataBuilder) RepositoryURL() string {
	remote, err := gm.repository.Remote("origin")
	if err != nil {
		return ""
	}

	remoteURLs := remote.Config().URLs
	if len(remoteURLs) > 0 {
		return remoteURLs[0]
	}

	return ""
}

func collectGitMetadata(repo *git.Repository) *gitMetadataBuilder {
	gm := &gitMetadataBuilder{
		repository: repo,
	}

	head, err := gm.repository.Head()
	if err != nil {
		util.Debugf("failed to resolve HEAD: %w", err)
		return gm
	}

	gm.commit = head

	if head.Name().IsBranch() {
		gm.branch = head.Name()
	}

	tags, err := gm.repository.Tags()
	if err != nil {
		util.Debugf("failed to read tags from repository: %w", err)
		return gm
	}

	ref, err := gm.FindReference(head.Hash(), tags)
	if err == nil {
		gm.tag = ref.Name()
	}

	return gm
}

func (gm *gitMetadataBuilder) Build() *GitMetadata {
	return &GitMetadata{
		Repository: gm.RepositoryURL(),
		Commit:     gm.CommitHash(),
		Branch:     gm.BranchName(),
		Tag:        gm.TagName(),
	}
}

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

	gitMetadata := collectGitMetadata(repositoryInfo.Repository).Build()
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
