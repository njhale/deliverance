package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ecordell/bndlr/pkg/registry"
	"github.com/ecordell/bndlr/pkg/registry/common"
	"github.com/ecordell/bndlr/pkg/registry/filestore"
	"github.com/ecordell/bndlr/pkg/registry/memory"
	"github.com/ecordell/bndlr/pkg/registry/store"
	"github.com/ecordell/bndlr/pkg/signals"
)

type StoreType string

const (
	MemoryStoreType StoreType = "memory"
	TmpFileStoreType StoreType = "tmp"
	FileStoreType StoreType = "file"
)

type pushOptions struct {
	// auth
	configs  []string
	username string
	password string

	storeType string
	storeDir string

	debug bool
}

var pushOpts pushOptions

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signals.Context()

		if len(args) < 2  {
			return fmt.Errorf("should be called with two args: dir host")
		}
		dir := args[0]
		ref := args[1]

		if pushOpts.debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		resolver := registry.NewResolver(pushOpts.username, pushOpts.password, pushOpts.configs...)
		var store store.Store
		var err error
		if pushOpts.storeType == string(MemoryStoreType) {
			store = memory.NewMemoryStore()
		} else if pushOpts.storeType == string(TmpFileStoreType) {
			store, err = filestore.NewTmpFileStore()
			if err != nil {
				return err
			}
		} else if pushOpts.storeType == string(FileStoreType) {
			if pushOpts.storeDir == "" {
				return fmt.Errorf("must specify --storagePath when using storage type file")
			}
			store, err = filestore.NewFileStore(pushOpts.storeDir)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("store type %s not supported", pushOpts.storeType)
		}

		digest, err := common.BuildAndPushDirectoryV22(ctx, ref, store, resolver, dir)
		if err != nil {
			return err
		}

		fmt.Printf("pushed with digest %s\n", digest.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringArrayVarP(&pushOpts.configs, "config", "c", []string{"~/.docker/config.json"}, "auth config path")
	pushCmd.Flags().StringVarP(&pushOpts.username, "username", "u", "", "username")
	pushCmd.Flags().StringVarP(&pushOpts.password, "password", "p", "", "password")
	pushCmd.Flags().BoolVarP(&pushOpts.debug, "debug", "d", false, "enable debug logging")
	pushCmd.Flags().StringVarP(&pushOpts.storeType, "storage", "s", string(TmpFileStoreType), "configure storage. Options: memory, tmp, file" )
	pushCmd.Flags().StringVar(&pushOpts.storeDir, "storagePath",  "", "configure storage location. only valid for storage type file" )
}
