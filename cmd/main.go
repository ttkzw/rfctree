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
		Name:  "rfctree",
		Usage: "Generate RFCs tree diagram",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "directory of RFCs' files",
				Value: "rfcs",
			},
			&cli.StringSliceFlag{
				Name:  "keyword",
				Usage: "keyword",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			dir := cmd.String("dir")
			keywords := cmd.StringSlice("keyword")
			return rfctree.GenerateTreeDiagram(dir, keywords)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
