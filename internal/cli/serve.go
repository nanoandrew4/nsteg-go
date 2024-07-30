package cli

import (
	"github.com/spf13/cobra"
	"nsteg/internal/server"
)

func ServeAppCommand() *cobra.Command {
	var port string

	command := &cobra.Command{
		Use:     "serve",
		Short:   "Serve an API to perform steganography over the web",
		Example: "nsteg serve --port 8888",
		Run: func(cmd *cobra.Command, args []string) {
			server.StartServer(port)
		},
	}

	command.Flags().StringVar(&port, "port", "8080", "Port on which to start the server")

	return command
}
