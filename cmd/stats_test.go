package cmd

import (
	"testing"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

const (
	invalidAppmap  = "this is not json"
	noEventsAppmap = `
	{
		"metadata": {},
		"classMap": []
	}
  `
)

func setup(appmap string) (*cobra.Command, string) {
	fs := afero.NewMemMapFs()
	config.SetFileSystem(fs)

	fname := "test.appmap.json"
	afero.WriteFile(fs, fname, []byte(appmap), 0755)
	afero.WriteFile(fs, "appmap.yml", []byte(appmapYml), 0755)

	return NewStatsCommand(&StatsProcessor{}), fname
}

func TestValidFile(t *testing.T) {
	cmd, fname := setup(validAppmap)

	assert.Nil(t, cmd.RunE(cmd, []string{fname}))
}

func TestInvalidFile(t *testing.T) {
	cmd, fname := setup(invalidAppmap)

	// Completely bogus file shouldn't fail the command
	assert.Nil(t, cmd.RunE(cmd, []string{fname}))
}

func TestMissingEvents(t *testing.T) {
	cmd, fname := setup(noEventsAppmap)

	// Missing events shouldn't fail the command
	assert.Nil(t, cmd.RunE(cmd, []string{fname}))
}

func TestFullAppmap(t *testing.T) {
	config.SetFileSystem(afero.NewOsFs())
	p := StatsProcessor{}
	appmap, err := p.ReadAppmap("testdata/test.appmap.json")
	assert.Nil(t, err)

	stats, count := p.MethodStats(appmap)
	assert.True(t, count == 60)
	requestStats := stats["Net::HTTP#request"]
	assert.NotNil(t, requestStats)
	assert.True(t, requestStats.Calls == 16)
	assert.True(t, len(requestStats.ParamCounts) == 8)
}
