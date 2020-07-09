package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	_ "github.com/mritd/goadmission/pkg/adfunc"
	"github.com/mritd/goadmission/pkg/conf"

	"github.com/mritd/goadmission/pkg/route"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var debug bool

var (
	version   string
	buildDate string
	commitID  string

	versionTpl = `
Name: goadmission
Version: %s
Arch: %s
BuildDate: %s
CommitID: %s
`
)

var rootCmd = &cobra.Command{
	Use:     "goadmission",
	Short:   "kubernetes dynamic admission control tool",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		router := route.Setup()
		srv := &http.Server{
			Handler: router,
			Addr:    conf.Addr,
		}

		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			var shutdownOnce sync.Once
			for range sigs {
				logrus.Warn("Receiving the termination signal, graceful shutdown...")
				shutdownOnce.Do(func() {
					err := srv.Shutdown(context.Background())
					if err != nil {
						logrus.Error(err)
					}
				})
			}
		}()

		if conf.Cert != "" && conf.Key != "" {
			err := srv.ListenAndServeTLS(conf.Cert, conf.Key)
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			err := srv.ListenAndServe()
			if err != nil {
				if err == http.ErrServerClosed {
					logrus.Info("server shutdown success")
				} else {
					logrus.Fatal(err)
				}
			}
		}
	},
}

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug level log")
	rootCmd.PersistentFlags().StringVarP(&conf.Addr, "listen", "l", ":443", "Admission Controller listen address")
	rootCmd.PersistentFlags().StringVar(&conf.Cert, "cert", "", "Admission Controller TLS cert")
	rootCmd.PersistentFlags().StringVar(&conf.Key, "key", "", "Admission Controller TLS cert key")
	rootCmd.PersistentFlags().StringSliceVar(&conf.ImageRename, "image-rename", conf.DefaultImageRenameRules, "Pod image name rename rules")
	rootCmd.SetVersionTemplate(fmt.Sprintf(versionTpl, version, runtime.GOOS+"/"+runtime.GOARCH, buildDate, commitID))
}

func initLog() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
