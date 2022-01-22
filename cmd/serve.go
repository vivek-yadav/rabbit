package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/vivek-yadav/rabbit/routes"
	"github.com/vivek-yadav/rabbit/settings"
	"github.com/vivek-yadav/rabbit/zlog"
	"go.uber.org/zap"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: serveRun,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serveRun(cmd *cobra.Command, args []string) {
	zap.S().Info("server called")
	fmt.Println("test:", settings.RunningConfig.Service.Port)
	if !strings.ContainsAny(settings.RunningConfig.Service.Port, ":") {
		settings.RunningConfig.Service.Port = ":" + settings.RunningConfig.Service.Port
	}
	if len(settings.RunningConfig.Service.Port) < 2 {
		err := errors.New("No Service.Port is set, cannot start the service")
		zlog.CheckFatal(err)
	}
	r := gin.New()

	logger, _ := zap.NewProduction()
	r.Use(zlog.Ginzap(logger, time.RFC3339, true))
	r.Use(zlog.RecoveryWithZap(logger, true))
	r.Use(zlog.RequestIdMiddleware())

	// Route Handlers / Endpoints
	routes.Routes(r)
	zap.S().Fatal(r.Run(settings.RunningConfig.Service.Port))
}
