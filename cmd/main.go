package main

import (
	"context"
	"log"
	"os"

	"github.com/ttkzw/rfctree"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:        "rfctree",
		Usage:       "Create RFCs tree diagram",
		Description: "Create RFCs tree diagram.",
		Commands: []*cli.Command{
			{
				Name:  "diagram",
				Usage: "Create RFCs tree diagram",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Usage:   "directory of RFCs' files",
						Value:   "rfcs",
					},
					&cli.StringFlag{
						Name:     "output",
						Usage:    "file name to output",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "keyword",
						Aliases: []string{"k"},
						Usage:   "keyword",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dir := cmd.String("dir")
					outputFilename := cmd.String("output")
					keywords := cmd.StringSlice("keyword")
					return rfctree.CreateDiagram(dir, outputFilename, keywords)
				},
			},
			{
				Name:  "keyword",
				Usage: "List related keywords",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Usage:   "directory of RFCs' files",
						Value:   "rfcs",
					},
					&cli.StringSliceFlag{
						Name:    "keyword",
						Aliases: []string{"k"},
						Usage:   "keyword",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dir := cmd.String("dir")
					keywords := cmd.StringSlice("keyword")
					return rfctree.ListKeyword(dir, keywords)
				},
			},
			{
				Name:  "list",
				Usage: "List RFCs",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Usage:   "directory of RFCs' files",
						Value:   "rfcs",
					},
					&cli.StringFlag{
						Name:     "output",
						Usage:    "file name to output",
						Required: false,
					},
					&cli.StringSliceFlag{
						Name:    "keyword",
						Aliases: []string{"k"},
						Usage:   "keyword",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dir := cmd.String("dir")
					outputFilename := cmd.String("output")
					keywords := cmd.StringSlice("keyword")
					return rfctree.List(dir, outputFilename, keywords)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
