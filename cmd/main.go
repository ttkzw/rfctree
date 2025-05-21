package main

import (
	"context"
	"log"
	"os"

	"github.com/ttkzw/rfctree"
	"github.com/ttkzw/rfctree/internal/cliutil"
	"github.com/urfave/cli/v3"
)

const (
	flagNameDir         = "dir"
	flagNameTarget      = "target"
	flagNameTargetFrom  = "target-from"
	flagNameKeyword     = "keyword"
	flagNameKeywordFrom = "keyword-from"
	flagNameExclude     = "exclude"
	flagNameExcludeFrom = "exclude-from"
	flagNameFollow      = "follow"
	flagNameOutput      = "output"
)

var (
	flagDir = &cli.StringFlag{
		Name:    flagNameDir,
		Aliases: []string{"d"},
		Usage:   "Directory containing RFC files.",
		Value:   "rfcs",
	}

	flagTarget = &cli.StringSliceFlag{
		Name:  flagNameTarget,
		Usage: "DocId, which format is 'RFC0821', specifying the target RFC. You can specify multiple times or separate them with a comma.",
	}

	flagTargetFrom = &cli.StringFlag{
		Name:  flagNameTargetFrom,
		Usage: "File specifying the target RFCs.",
	}

	flagKeyword = &cli.StringSliceFlag{
		Name:    flagNameKeyword,
		Aliases: []string{"k"},
		Usage:   "Keywords. You can specify multiple times or separate them with a comma.",
	}

	flagKeywordFrom = &cli.StringFlag{
		Name:  flagNameKeywordFrom,
		Usage: "File specifying keywords.",
	}

	flagExclude = &cli.StringSliceFlag{
		Name:  flagNameExclude,
		Usage: "DocId, which format is 'RFC0821', specifying RFC to be excluded. You can specify multiple times or separate them with a comma.",
	}

	flagExcludeFrom = &cli.StringFlag{
		Name:  flagNameExcludeFrom,
		Usage: "File specifying RFCs to be excluded.",
	}

	flagFollow = &cli.BoolFlag{
		Name:  flagNameFollow,
		Usage: "Follow the relations between RFCs.",
		Value: false,
	}

	flagOutput = &cli.StringFlag{
		Name:  flagNameOutput,
		Usage: "file to output.",
		Value: "rfctree.pdf",
	}
)

func main() {
	cmd := &cli.Command{
		Name:        "rfctree",
		Usage:       "Create RFCs tree diagram",
		Description: "Create RFCs tree diagram.",
		Flags:       []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:  "diagram",
				Usage: "Create RFCs tree diagram",
				Flags: []cli.Flag{
					flagDir,
					flagTarget,
					flagTargetFrom,
					flagKeyword,
					flagKeywordFrom,
					flagFollow,
					flagExclude,
					flagExcludeFrom,
					flagOutput,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					targets, err := cliutil.GetTargets(cmd.String(flagNameTargetFrom), cmd.StringSlice(flagNameTarget))
					if err != nil {
						return err
					}

					keywords, err := cliutil.GetKeywords(cmd.String(flagNameKeywordFrom), cmd.StringSlice(flagNameKeyword))
					if err != nil {
						return err
					}

					excludes, err := cliutil.GetExcludes(cmd.String(flagNameExcludeFrom), cmd.StringSlice(flagNameExclude))
					if err != nil {
						return err
					}

					return rfctree.CreateDiagram(cmd.String(flagNameDir), targets, keywords, cmd.Bool(flagNameFollow), excludes, cmd.String(flagNameOutput))
				},
			},
			{
				Name:  "keyword",
				Usage: "List related keywords",
				Flags: []cli.Flag{
					flagDir,
					flagTarget,
					flagTargetFrom,
					flagKeyword,
					flagKeywordFrom,
					flagFollow,
					flagExclude,
					flagExcludeFrom,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					targets, err := cliutil.GetTargets(cmd.String(flagNameTargetFrom), cmd.StringSlice(flagNameTarget))
					if err != nil {
						return err
					}

					keywords, err := cliutil.GetKeywords(cmd.String(flagNameKeywordFrom), cmd.StringSlice(flagNameKeyword))
					if err != nil {
						return err
					}

					excludes, err := cliutil.GetExcludes(cmd.String(flagNameExcludeFrom), cmd.StringSlice(flagNameExclude))
					if err != nil {
						return err
					}

					return rfctree.ListKeyword(cmd.String(flagNameDir), targets, keywords, cmd.Bool(flagNameFollow), excludes)
				},
			},
			{
				Name:  "list",
				Usage: "List RFCs",
				Flags: []cli.Flag{
					flagDir,
					flagTarget,
					flagTargetFrom,
					flagKeyword,
					flagKeywordFrom,
					flagFollow,
					flagExclude,
					flagExcludeFrom,
					flagOutput,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					targets, err := cliutil.GetTargets(cmd.String(flagNameTargetFrom), cmd.StringSlice(flagNameTarget))
					if err != nil {
						return err
					}

					keywords, err := cliutil.GetKeywords(cmd.String(flagNameKeywordFrom), cmd.StringSlice(flagNameKeyword))
					if err != nil {
						return err
					}

					excludes, err := cliutil.GetExcludes(cmd.String(flagNameExcludeFrom), cmd.StringSlice(flagNameExclude))
					if err != nil {
						return err
					}

					return rfctree.List(cmd.String(flagNameDir), targets, keywords, cmd.Bool(flagNameFollow), excludes, cmd.String(flagNameOutput))
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
