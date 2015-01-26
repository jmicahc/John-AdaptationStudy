// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package contains the main entry point of gd.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/odeke-em/drive/config"
	"github.com/odeke-em/drive/src"
	"github.com/rakyll/command"
)

var context *config.Context
var DefaultMaxProcs = runtime.NumCPU()

func main() {
	maxProcs, err := strconv.ParseInt(os.Getenv("GOMAXPROCS"), 10, 0)
	if err != nil || maxProcs < 1 {
		maxProcs = int64(DefaultMaxProcs)
	}
	runtime.GOMAXPROCS(int(maxProcs))

	command.On(drive.AboutKey, drive.DescAbout, &aboutCmd{}, []string{})
	command.On(drive.DiffKey, drive.DescDiff, &diffCmd{}, []string{})
	command.On(drive.EmptyTrashKey, drive.DescEmptyTrash, &emptyTrashCmd{}, []string{})
	command.On(drive.FeaturesKey, drive.DescFeatures, &featuresCmd{}, []string{})
	command.On(drive.InitKey, drive.DescInit, &initCmd{}, []string{})
	command.On(drive.HelpKey, drive.DescHelp, &helpCmd{}, []string{})
	command.On(drive.ListKey, drive.DescList, &listCmd{}, []string{})
	command.On(drive.PullKey, drive.DescPull, &pullCmd{}, []string{})
	command.On(drive.PushKey, drive.DescPush, &pushCmd{}, []string{})
	command.On(drive.PubKey, drive.DescPublish, &publishCmd{}, []string{})
	command.On(drive.QuotaKey, drive.DescQuota, &quotaCmd{}, []string{})
	command.On(drive.TouchKey, drive.DescTouch, &touchCmd{}, []string{})
	command.On(drive.TrashKey, drive.DescTrash, &trashCmd{}, []string{})
	command.On(drive.UntrashKey, drive.DescUntrash, &untrashCmd{}, []string{})
	command.On(drive.UnpubKey, drive.DescUnpublish, &unpublishCmd{}, []string{})
	command.On(drive.VersionKey, drive.Version, &versionCmd{}, []string{})
	command.ParseAndRun()
}

type helpCmd struct {
	args []string
}

func (cmd *helpCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *helpCmd) Run(args []string) {
	if len(args) < 1 {
		exitWithError(fmt.Errorf("help for more usage"))
	}
	drive.ShowDescription(args[0])
	exitWithError(nil)
}

type featuresCmd struct{}

func (cmd *featuresCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *featuresCmd) Run(args []string) {
	context, path := discoverContext(args)
	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).About(drive.AboutFeatures))
}

type versionCmd struct{}

func (cmd *versionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *versionCmd) Run(args []string) {
	drive.PrintVersion()
	exitWithError(nil)
}

type initCmd struct{}

func (cmd *initCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *initCmd) Run(args []string) {
	exitWithError(drive.New(initContext(args), nil).Init())
}

type quotaCmd struct{}

func (cmd *quotaCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *quotaCmd) Run(args []string) {
	context, path := discoverContext(args)
	exitWithError(drive.New(context, &drive.Options{
		Path: path,
	}).About(drive.AboutQuota))
}

type listCmd struct {
	hidden      *bool
	pageCount   *int
	recursive   *bool
	files       *bool
	directories *bool
	depth       *int
	pageSize    *int64
	longFmt     *bool
	noPrompt    *bool
	inTrash     *bool
}

func (cmd *listCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.depth = fs.Int("m", 1, "maximum recursion depth")
	cmd.hidden = fs.Bool("hidden", false, "list all paths even hidden ones")
	cmd.files = fs.Bool("f", false, "list only files")
	cmd.directories = fs.Bool("d", false, "list all directories")
	cmd.longFmt = fs.Bool("l", false, "long listing of contents")
	cmd.pageSize = fs.Int64("p", 100, "number of results per pagination")
	cmd.inTrash = fs.Bool("trashed", false, "list content in the trash")
	cmd.noPrompt = fs.Bool("no-prompt", false, "shows no prompt before pagination")
	cmd.recursive = fs.Bool("r", false, "recursively list subdirectories")

	return fs
}

func (cmd *listCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)

	typeMask := 0
	if *cmd.directories {
		typeMask |= drive.Folder
	}
	if *cmd.files {
		typeMask |= drive.NonFolder
	}
	if *cmd.inTrash {
		typeMask |= drive.InTrash
	}
	if !*cmd.longFmt {
		typeMask |= drive.Minimal
	}

	exitWithError(drive.New(context, &drive.Options{
		Depth:     *cmd.depth,
		Hidden:    *cmd.hidden,
		InTrash:   *cmd.inTrash,
		PageSize:  *cmd.pageSize,
		Path:      path,
		NoPrompt:  *cmd.noPrompt,
		Recursive: *cmd.recursive,
		Sources:   sources,
		TypeMask:  typeMask,
	}).List())
}

type pullCmd struct {
	exportsDir *string
	export     *string
	force      *bool
	hidden     *bool
	noPrompt   *bool
	noClobber  *bool
	recursive  *bool
}

func (cmd *pullCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noClobber = fs.Bool("no-clobber", false, "prevents overwriting of old content")
	cmd.export = fs.String(
		"export", "", "comma separated list of formats to export your docs + sheets files")
	cmd.recursive = fs.Bool("r", true, "performs the pull action recursively")
	cmd.noPrompt = fs.Bool("no-prompt", false, "shows no prompt before applying the pull action")
	cmd.hidden = fs.Bool("hidden", false, "allows pulling of hidden paths")
	cmd.force = fs.Bool("force", false, "forces a pull even if no changes present")
	cmd.exportsDir = fs.String("export-dir", "", "directory to place exports")

	return fs
}

func nonEmptyStrings(v []string) (splits []string) {
	for _, elem := range v {
		if elem != "" {
			splits = append(splits, elem)
		}
	}
	return
}

func (cmd *pullCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)

	// Filter out empty strings.
	exports := nonEmptyStrings(strings.Split(*cmd.export, ","))

	exitWithError(drive.New(context, &drive.Options{
		Exports:    uniqOrderedStr(exports),
		ExportsDir: strings.Trim(*cmd.exportsDir, " "),
		Force:      *cmd.force,
		Hidden:     *cmd.hidden,
		NoPrompt:   *cmd.noPrompt,
		NoClobber:  *cmd.noClobber,
		Path:       path,
		Recursive:  *cmd.recursive,
		Sources:    sources,
	}).Pull())
}

type pushCmd struct {
	noClobber   *bool
	hidden      *bool
	force       *bool
	noPrompt    *bool
	recursive   *bool
	mountedPush *bool
}

func (cmd *pushCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noClobber = fs.Bool("no-clobber", false, "allows overwriting of old content")
	cmd.hidden = fs.Bool("hidden", false, "allows pushing of hidden paths")
	cmd.recursive = fs.Bool("r", true, "performs the push action recursively")
	cmd.noPrompt = fs.Bool("no-prompt", false, "shows no prompt before applying the push action")
	cmd.force = fs.Bool("force", false, "forces a push even if no changes present")
	cmd.mountedPush = fs.Bool("m", false, "allows pushing of mounted paths")
	return fs
}

func preprocessArgs(args []string) ([]string, *config.Context, string) {
	var relPaths []string
	context, path := discoverContext(args)
	root := context.AbsPathOf("")

	if len(args) < 1 {
		args = []string{"."}
	}

	var err error
	for _, p := range args {
		p, err = filepath.Abs(p)
		if err != nil {
			fmt.Println(err)
			continue
		}

		relPath, err := filepath.Rel(root, p)
		if relPath == "." {
			relPath = ""
		}

		exitWithError(err)

		relPath = "/" + relPath
		relPaths = append(relPaths, relPath)
	}

	return uniqOrderedStr(relPaths), context, path
}

func (cmd *pushCmd) Run(args []string) {
	if *cmd.mountedPush {
		pushMounted(cmd, args)
	} else {
		sources, context, path := preprocessArgs(args)
		exitWithError(drive.New(context, &drive.Options{
			Force:     *cmd.force,
			Hidden:    *cmd.hidden,
			NoClobber: *cmd.noClobber,
			NoPrompt:  *cmd.noPrompt,
			Path:      path,
			Recursive: *cmd.recursive,
			Sources:   sources,
		}).Push())
	}
}

type touchCmd struct {
	hidden    *bool
	noPrompt  *bool
	recursive *bool
}

func (cmd *touchCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool("hidden", false, "allows pushing of hidden paths")
	cmd.recursive = fs.Bool("r", true, "performs the push action recursively")
	cmd.noPrompt = fs.Bool("no-prompt", false, "shows no prompt before applying the push action")
	return fs
}

func (cmd *touchCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		Hidden:    *cmd.hidden,
		NoPrompt:  *cmd.noPrompt,
		Path:      path,
		Recursive: *cmd.recursive,
		Sources:   sources,
	}).Touch())
}

func pushMounted(cmd *pushCmd, args []string) {
	argc := len(args)

	var contextArgs, rest, sources []string

	if !*cmd.mountedPush {
		contextArgs = args
	} else {
		// Expectation is that at least one path has to be passed
		if argc < 2 {
			cwd, cerr := os.Getwd()
			if cerr != nil {
				contextArgs = []string{cwd}
			}
			rest = args
		} else {
			rest = args[:argc-1]
			contextArgs = args[argc-1:]
		}
	}

	rest = nonEmptyStrings(rest)
	context, path := discoverContext(contextArgs)
	contextAbsPath, err := filepath.Abs(path)
	if path == "." {
		path = ""
	}
	exitWithError(err)

	mountPoints, auxSrcs := config.MountPoints(path, contextAbsPath, rest, *cmd.hidden)
	sources = append(sources, auxSrcs...)

	exitWithError(drive.New(context, &drive.Options{
		Hidden:    *cmd.hidden,
		NoPrompt:  *cmd.noPrompt,
		Recursive: *cmd.recursive,
		Mounts:    mountPoints,
		NoClobber: *cmd.noClobber,
		Path:      path,
		Sources:   sources,
	}).Push())
}

type aboutCmd struct {
	features *bool
	quota    *bool
	filesize *bool
}

func (cmd *aboutCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.features = fs.Bool("features", false, "gives information on features present on this drive")
	cmd.quota = fs.Bool("quota", false, "prints out quota information for this drive")
	cmd.filesize = fs.Bool("filesize", false, "prints out information about file sizes e.g the max upload size for a specific file size")
	return fs
}

func (cmd *aboutCmd) Run(args []string) {
	_, context, _ := preprocessArgs(args)

	mask := drive.AboutNone
	if *cmd.features {
		mask |= drive.AboutFeatures
	}
	if *cmd.quota {
		mask |= drive.AboutQuota
	}
	if *cmd.filesize {
		mask |= drive.AboutFileSizes
	}
	exitWithError(drive.New(context, &drive.Options{}).About(mask))
}

type diffCmd struct {
	hidden *bool
}

func (cmd *diffCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool("hidden", false, "allows pulling of hidden paths")
	return fs
}

func (cmd *diffCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		Recursive: true,
		Path:      path,
		Hidden:    *cmd.hidden,
		Sources:   sources,
	}).Diff())
}

type publishCmd struct {
	hidden *bool
}

type unpublishCmd struct {
	hidden *bool
}

func (cmd *unpublishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool("hidden", false, "allows pulling of hidden paths")
	return fs
}

func (cmd *unpublishCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
	}).Unpublish())
}

type emptyTrashCmd struct {
	noPrompt *bool
}

func (cmd *emptyTrashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.noPrompt = fs.Bool("no-prompt", false, "shows no prompt before emptying the trash")
	return fs
}

func (cmd *emptyTrashCmd) Run(args []string) {
	_, context, _ := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		NoPrompt: *cmd.noPrompt,
	}).EmptyTrash())
}

type trashCmd struct {
	hidden *bool
}

func (cmd *trashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool("hidden", false, "allows trashing hidden paths")
	return fs
}

func (cmd *trashCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
	}).Trash())
}

type untrashCmd struct {
	hidden *bool
}

func (cmd *untrashCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool("hidden", false, "allows untrashing hidden paths")
	return fs
}

func (cmd *untrashCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
	}).Untrash())
}

func (cmd *publishCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.hidden = fs.Bool("hidden", false, "allows publishing of hidden paths")
	return fs
}

func (cmd *publishCmd) Run(args []string) {
	sources, context, path := preprocessArgs(args)
	exitWithError(drive.New(context, &drive.Options{
		Path:    path,
		Sources: sources,
	}).Publish())
}

func initContext(args []string) *config.Context {
	var err error
	var gdPath string
	var firstInit bool

	gdPath, firstInit, context, err = config.Initialize(getContextPath(args))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// The signal handler should clean up the .gd path if this is the first time
	go func() {
		_ = <-c
		if firstInit {
			os.RemoveAll(gdPath)
		}
		os.Exit(1)
	}()

	exitWithError(err)
	return context
}

func discoverContext(args []string) (*config.Context, string) {
	var err error
	context, err = config.Discover(getContextPath(args))
	exitWithError(err)
	relPath := ""
	if len(args) > 0 {
		var headAbsArg string
		headAbsArg, err = filepath.Abs(args[0])
		if err == nil {
			relPath, err = filepath.Rel(context.AbsPath, headAbsArg)
		}
	}

	exitWithError(err)

	// relPath = strings.Join([]string{"", relPath}, "/")
	return context, relPath
}

func getContextPath(args []string) (contextPath string) {
	if len(args) > 0 {
		contextPath, _ = filepath.Abs(args[0])
	}
	if contextPath == "" {
		contextPath, _ = os.Getwd()
	}
	return
}

func uniqOrderedStr(sources []string) []string {
	cache := map[string]bool{}
	var uniqPaths []string
	for _, p := range sources {
		ok := cache[p]
		if ok {
			continue
		}
		uniqPaths = append(uniqPaths, p)
		cache[p] = true
	}
	return uniqPaths
}

func exitWithError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
