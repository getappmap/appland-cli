package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/applandinc/appland-cli/internal/files"
	"github.com/spf13/cobra"
)

type StatsProcessor struct {
	verbose bool
	files   bool
	params  bool
	limit   int
	json    bool
}

type param struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type sqlQuery struct {
	SQL          string `json:"sql"`
	DatabaseType string `json:"database_type"`
}

type event struct {
	Event        string   `json:"event"`
	DefinedClass string   `json:"defined_class"`
	MethodID     string   `json:"method_id"`
	Static       bool     `json:"static"`
	Parameters   []param  `json:"parameters"`
	SQLQuery     sqlQuery `json:"sql_query"`
}

type Appmap struct {
	Events []event `json:"events"`
}

type Stats struct {
	processor   StatsProcessor
	Calls       int            `json:"calls"`
	NumParams   int            `json:"num_params"`
	ParamCounts map[string]int `json:"param_counts"`
}

type total struct {
	Method string `json:"method"`
	Stats
}

func (t total) MarshalJSON() ([]byte, error) {
	var v struct {
		Method      string          `json:"method"`
		Calls       int             `json:"calls"`
		NumParams   *int            `json:"num_params,omitempty"`
		ParamCounts *map[string]int `json:"param_counts,omitempty"`
	}

	v.Method = t.Method
	v.Calls = t.Stats.Calls
	if t.processor.params {
		v.NumParams = &t.Stats.NumParams
		v.ParamCounts = &t.Stats.ParamCounts
	}

	return json.Marshal(v)
}

func (p StatsProcessor) sortStatsByCount(stats map[string]Stats) []total {
	var t []total
	for k, v := range stats {
		t = append(t, total{k, Stats{processor: p, Calls: v.Calls, NumParams: v.NumParams, ParamCounts: v.ParamCounts}})
	}

	sort.Slice(t, func(i, j int) bool {
		// Sort by number of calls
		ret := t[i].Stats.Calls > t[j].Stats.Calls
		// then by method name
		if t[i].Stats.Calls == t[j].Stats.Calls {
			ret = t[i].Method < t[j].Method
		}
		return ret
	})

	return t
}

func countEvents(appmap Appmap) int {
	if appmap.Events != nil {
		return len(appmap.Events)
	}

	return 0
}

func (p StatsProcessor) ReadAppmap(fname string) (Appmap, error) {
	fs := config.GetFS()
	f, err := fs.Open(fname)
	if err != nil {
		return Appmap{}, fmt.Errorf("Failed opening %s: %w", fname, err)
	} else if p.verbose {
		fmt.Fprintf(os.Stderr, "Processing %s\n", fname)
	}

	dec := json.NewDecoder(f)
	var appmap Appmap
	err = dec.Decode(&appmap)

	f.Close()
	if err != nil {
		return Appmap{}, fmt.Errorf(">>> Failed decoding %s, %w", fname, err)
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "%s: %d event(s)\n", fname, countEvents(appmap))
	}

	return appmap, nil
}

func (p StatsProcessor) MethodStats(appmap Appmap) (map[string]Stats, uint64) {
	stats := make(map[string]Stats)
	if countEvents(appmap) == 0 {
		return stats, 0
	}

	var total uint64

	for _, event := range appmap.Events {
		if event.Event != "call" || event.DefinedClass == "" {
			continue
		}
		sep := "#"
		if event.Static {
			sep = "."
		}
		id := strings.Join([]string{event.DefinedClass, sep, event.MethodID}, "")
		params := event.Parameters
		var (
			paramId  string
			paramSep string
		)

		for _, param := range params {
			var v string
			if param.Value != nil {
				v = param.Value.(string)
			} else {
				v = "<nil>"
			}
			paramId += paramSep + v
			paramSep = ","
		}

		if _, ok := stats[id]; !ok {
			stats[id] = Stats{processor: p, NumParams: len(params), ParamCounts: make(map[string]int)}
		}
		newStats := stats[id]
		newStats.Calls++
		stats[id] = newStats
		stats[id].ParamCounts[paramId]++
		total++
	}

	return stats, total
}

func (p StatsProcessor) printStats(msg string, calls uint64, methodStats map[string]Stats) {
	totals := p.sortStatsByCount(methodStats)

	max := p.limit
	if len(totals) < max {
		max = len(totals)
	}
	totals = totals[0:max]

	if msg != "" {
		fmt.Print(msg + ": ")
	}
	if p.json {
		j, err := json.MarshalIndent(totals, "", "  ")
		if err == nil {
			fmt.Print(string(j))
		} else {
			warn(err)
		}
	} else {
		fmt.Printf("%d calls, top %d methods\n", calls, max)
		for _, t := range totals {
			key := t.Method
			distinct := len(methodStats[key].ParamCounts)
			fmt.Printf("  %s: %d (%d distinct)\n", key, t.Calls, distinct)
			if p.params {
				hasParams := methodStats[key].NumParams > 0
				if hasParams {
					fmt.Println("   has parameters")
					for params, _ := range methodStats[key].ParamCounts {
						fmt.Printf("    %v\n", params)
					}
				} else {
					fmt.Println("   no parameters")
				}
			}
		}
	}
}

func NewStatsCommand(p *StatsProcessor) *cobra.Command {
	return &cobra.Command{
		Use:   "stats [files, directories]",
		Short: "Show statistics for AppMaps",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnames, err := files.FindAppMaps(args)
			if err != nil {
				return fmt.Errorf("Failed finding AppMaps: %w", err)
			}

			var (
				totalMethodCalls   uint64 = 0
				globalMethodCounts        = make(map[string]Stats)
			)

			if p.verbose {
				fmt.Fprintf(os.Stderr, "Found %d appmap(s)\n", len(fnames))
			}

			for _, fname := range fnames {
				appmap, err := p.ReadAppmap(fname)
				if err != nil {
					warn(err)
					continue
				}

				if appmap.Events == nil {
					if p.verbose {
						fmt.Fprintf(os.Stderr, "%s, events is nil\n", fname)
					}
					continue
				}

				methodStats, calls := p.MethodStats(appmap)

				if p.files {
					p.printStats(fname, calls, methodStats)
				}

				for id, stats := range methodStats {
					if _, ok := globalMethodCounts[id]; !ok {
						globalMethodCounts[id] = Stats{processor: *p, NumParams: stats.NumParams, ParamCounts: make(map[string]int)}
					}
					newStats := globalMethodCounts[id]
					newStats.Calls += stats.Calls
					for param, count := range stats.ParamCounts {
						newStats.ParamCounts[param] += count
					}
					globalMethodCounts[id] = newStats
				}
				totalMethodCalls += calls
			}

			if p.files {
				fmt.Print("\n\n")
			}

			p.printStats("", totalMethodCalls, globalMethodCounts)

			return nil
		},
	}
}

func init() {
	var (
		processor = StatsProcessor{}
		statsCmd  = NewStatsCommand(&processor)
	)

	flags := statsCmd.Flags()
	flags.BoolVarP(&processor.verbose, "verbose", "v", false, "be verbose while processing")
	flags.BoolVarP(&processor.files, "files", "f", false, "show statistics for each file")
	flags.BoolVarP(&processor.params, "params", "p", false, "show distinct parameters for each method")
	flags.IntVarP(&processor.limit, "limit", "l", 20, "limit the number of methods displayed")
	flags.BoolVarP(&processor.json, "json", "j", false, "format results as JSON")

	rootCmd.AddCommand(statsCmd)
}
