package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/mritd/goadmission/pkg/adfunc"

	"github.com/mritd/goadmission/pkg/zaplogger"

	"github.com/mritd/goadmission/pkg/conf"

	"github.com/mritd/goadmission/pkg/route"
	"github.com/spf13/cobra"
)

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
		zaplogger.Setup()
		adfunc.Setup()
		route.Setup()

		logger := zaplogger.NewSugar("main")

		srv := &http.Server{
			Handler: route.Router(),
			Addr:    conf.Addr,
		}

		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			var shutdownOnce sync.Once
			for range sigs {
				logger.Warn("Receiving the termination signal, graceful shutdown...")
				shutdownOnce.Do(func() {
					err := srv.Shutdown(context.Background())
					if err != nil {
						logger.Error(err)
					}
				})
			}
		}()

		if conf.Cert != "" && conf.Key != "" {
			logger.Infof("Listen TLS Server at %s", conf.Addr)
			err := srv.ListenAndServeTLS(conf.Cert, conf.Key)
			if err != nil {
				if err == http.ErrServerClosed {
					logger.Info("server shutdown success.")
				} else {
					logger.Fatal(err)
				}
			}
		} else {
			logger.Infof("Listen HTTP Server at %s", conf.Addr)
			err := srv.ListenAndServe()
			if err != nil {
				if err == http.ErrServerClosed {
					logger.Info("server shutdown success.")
				} else {
					logger.Fatal(err)
				}
			}
		}
	},
}

func init() {
	// zap logger
	rootCmd.PersistentFlags().BoolVar(&zaplogger.Config.Development, "zap-devel", false, "Enable zap development mode (changes defaults to console encoder, debug log level, disables sampling and stacktrace from 'warning' level)")
	rootCmd.PersistentFlags().StringVar(&zaplogger.Config.Encoder, "zap-encoder", "console", "Zap log encoding ('json' or 'console')")
	rootCmd.PersistentFlags().StringVar(&zaplogger.Config.Level, "zap-level", "info", "Zap log level (one of 'debug', 'info', 'warn', 'error')")
	rootCmd.PersistentFlags().BoolVar(&zaplogger.Config.Sample, "zap-sample", false, "Enable zap log sampling. Sampling will be disabled for log level is debug")
	rootCmd.PersistentFlags().StringVar(&zaplogger.Config.TimeEncoding, "zap-time-encoding", "default", "Sets the zap time format ('default', 'epoch', 'millis', 'nano', or 'iso8601')")
	rootCmd.PersistentFlags().StringVar(&zaplogger.Config.StackLevel, "zap-stacktrace-level", "error", "Set the minimum log level that triggers stacktrace generation")

	// version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(versionTpl, version, runtime.GOOS+"/"+runtime.GOARCH, buildDate, commitID))

	// webhook
	rootCmd.PersistentFlags().StringVarP(&conf.Addr, "listen", "l", ":443", "Admission Controller listen address")
	rootCmd.PersistentFlags().StringVar(&conf.Cert, "cert", "", "Admission Controller TLS cert")
	rootCmd.PersistentFlags().StringVar(&conf.Key, "key", "", "Admission Controller TLS cert key")

	// adfunc image_rename
	rootCmd.PersistentFlags().StringSliceVar(&conf.ImageRename, "image-rename", conf.DefaultImageRenameRules, "Pod image name rename rules")
	// adfunc check_deploy_time
	rootCmd.PersistentFlags().StringSliceVar(&conf.AllowDeployTime, "allow-deploy-time", conf.DefaultAllowDeployTime, "Allow deploy time")
	rootCmd.PersistentFlags().StringVar(&conf.ForceDeployLabel, "force-deploy-label", conf.DefaultForceDeployLabel, "Force deploy label")
	// adfunc disable_service_links
	rootCmd.PersistentFlags().StringVar(&conf.ForceEnableServiceLinksLabel, "force-enable-service-links-label", conf.DefaultForceEnableServiceLinksLabel, "Force enable service links label")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
