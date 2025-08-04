package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apiServer "mira/cmd/api"
	imageBuilder "mira/cmd/image-builder"
)

var APIServerCmd = &cobra.Command{
	Use:   "api-server",
	Short: "Start the MIRA API Server",
	Long: `Start the MIRA API Server.
`,
	Run: func(cmd *cobra.Command, args []string) {
		port := cmd.Flag("port").Value.String()
		if port == "" {
			port = "3000"
		}
		fmt.Println("Starting MIRA API Server on port:", port)
		apiServer.StartServer(port)
		fmt.Println("MIRA API Server started successfully.")
	},
}

var ImageBuilderCmd = &cobra.Command{
	Use:   "image-builder",
	Short: "Start the MIRA Image Builder",
	Long: `Start the MIRA Image Builder.
`,
	Run: func(cmd *cobra.Command, args []string) {
		imageBuilder.Listen()
	},
}

func init() {
	APIServerCmd.Flags().StringP("port", "p", "", "Port to run the API Server on (default: 3000)")
}
