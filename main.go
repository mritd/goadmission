package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mritd/goadmission/pkg/adfunc"
	"github.com/mritd/goadmission/pkg/conf"

	"github.com/mritd/goadmission/pkg/route"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "goadmission",
	Short: "kubernetes dynamic admission control tool",
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
	rootCmd.PersistentFlags().StringVarP(&conf.Addr, "listen", "l", ":8080", "Admission Controller listen address")
	rootCmd.PersistentFlags().StringVar(&conf.Cert, "cert", "", "Admission Controller TLS cert")
	rootCmd.PersistentFlags().StringVar(&conf.Key, "key", "", "Admission Controller TLS cert key")
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
