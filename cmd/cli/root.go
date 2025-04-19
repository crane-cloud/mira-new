package cli

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var Version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mira",
	Short: "Auto-containerazation of source code",
	Long: `Auto-containerazation of source code
	`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Mira")
		fmt.Println("Use mira --help to see available commands")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	rootCmd.AddCommand(APIServerCmd)
	rootCmd.AddCommand(ImageBuilderCmd)
	// Here you will define your flags and configuration settings.
}
