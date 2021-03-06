package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"path"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsGetCmd.Flags().StringP("token", "t", "", "token of the request")
	ffsGetCmd.Flags().String("ipfsrevproxy", "localhost:6002", "Powergate IPFS reverse proxy DNS address. If port 443, is assumed is a HTTPS endpoint.")
	ffsGetCmd.Flags().BoolP("folder", "f", false, "Indicates that the retrieved Cid is a folder")
	ffsCmd.AddCommand(ffsGetCmd)
}

var ffsGetCmd = &cobra.Command{
	Use:   "get [cid] [output file path]",
	Short: "Get data by cid from ffs",
	Long:  `Get data by cid from ffs`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		if len(args) != 2 {
			Fatal(errors.New("you must provide cid and output file path arguments"))
		}

		c, err := cid.Parse(args[0])
		checkErr(err)

		s := spin.New("%s Retrieving specified data...")
		s.Start()

		isFolder := viper.GetBool("folder")
		if isFolder {
			err := fcClient.FFS.GetFolder(authCtx(ctx), viper.GetString("ipfsrevproxy"), c, args[1])
			checkErr(err)
		} else {
			reader, err := fcClient.FFS.Get(authCtx(ctx), c)
			checkErr(err)

			dir := path.Dir(args[1])
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.MkdirAll(dir, os.ModePerm)
				checkErr(err)
			}
			file, err := os.Create(args[1])
			checkErr(err)

			defer func() { checkErr(file.Close()) }()

			buffer := make([]byte, 1024*32) // 32KB
			for {
				bytesRead, readErr := reader.Read(buffer)
				if readErr != nil && readErr != io.EOF {
					Fatal(readErr)
				}
				_, err = file.Write(buffer[:bytesRead])
				checkErr(err)
				if readErr == io.EOF {
					break
				}
			}
		}
		s.Stop()
		Success("Data written to %v", args[1])
	},
}
