package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/internal/app/pimo"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/add"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/command"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/constant"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/dateparser"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/duration"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/fluxuri"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/hash"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/increment"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/jsonline"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/model"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/randdate"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/randdura"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/randomdecimal"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/randomint"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/randomlist"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/rangemask"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/regex"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/remove"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/replacement"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/templatemask"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/weightedchoice"
)

// Provisioned by ldflags
// nolint: gochecknoglobals
var (
	version   string
	commit    string
	buildDate string
	builtBy   string

	iteration    int
	emptyInput   bool
	maskingFile  string
	cachesToDump map[string]string
	cachesToLoad map[string]string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "pimo",
		Short:   "Command line to mask data from jsonlines",
		Long:    `Pimo is a tool to mask private data contained in jsonlines by using masking configurations`,
		Version: fmt.Sprintf("%v (commit=%v date=%v by=%v)\n© CGI Inc 2020 All rights reserved", version, commit, buildDate, builtBy),
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	rootCmd.PersistentFlags().IntVarP(&iteration, "repeat", "r", 1, "number of iteration to mask each input")
	rootCmd.PersistentFlags().BoolVar(&emptyInput, "empty-input", false, "generate datas without any input, to use with repeat flag")
	rootCmd.PersistentFlags().StringVarP(&maskingFile, "config", "c", "masking.yml", "name and location of the masking-config file")
	rootCmd.PersistentFlags().StringToStringVar(&cachesToDump, "dump-cache", map[string]string{}, "path for dumping cache into file")
	rootCmd.PersistentFlags().StringToStringVar(&cachesToLoad, "load-cache", map[string]string{}, "path for loading cache from file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() {
	var source model.ISource
	if emptyInput {
		source = model.NewSourceFromSlice([]model.Dictionary{{}})
	} else {
		source = jsonline.NewSource(os.Stdin)
	}
	pipeline := model.NewPipeline(source).
		Process(model.NewRepeaterProcess(iteration))

	var (
		err    error
		caches map[string]model.ICache
	)

	pipeline, caches, err = pimo.YamlPipeline(pipeline, maskingFile, injectMaskFactories(), injectMaskContextFactories())

	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	for name, path := range cachesToLoad {
		cache, ok := caches[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "Cache %s not found", name)
			os.Exit(2)
		}
		err = pimo.LoadCache(name, cache, path)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(3)
		}
	}

	err = pipeline.AddSink(jsonline.NewSink(os.Stdout)).Run()

	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(4)
	}

	for name, path := range cachesToDump {
		cache, ok := caches[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "Cache %s not found", name)
			os.Exit(2)
		}
		err = pimo.DumpCache(name, cache, path)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(3)
		}
	}

	os.Exit(0)
}

func injectMaskContextFactories() []model.MaskContextFactory {
	return []model.MaskContextFactory{
		fluxuri.Factory,
		add.Factory,
		remove.Factory,
	}
}

func injectMaskFactories() []model.MaskFactory {
	return []model.MaskFactory{

		constant.Factory,
		command.Factory,
		randomlist.Factory,
		randomint.Factory,
		weightedchoice.Factory,
		regex.Factory,
		hash.Factory,
		randdate.Factory,
		increment.Factory,
		replacement.Factory,
		duration.Factory,
		templatemask.Factory,

		rangemask.Factory,
		randdura.Factory,
		randomdecimal.Factory,
		dateparser.Factory,
	}
}
