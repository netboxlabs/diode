package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/netboxlabs/diode/diode-server/auth"
	"github.com/netboxlabs/diode/diode-server/authutil"
)

// SubCommand is an interface for a cli subcommand.
type SubCommand interface {
	Run() error
}

const (
	defaultTimeout = 10 * time.Second
)

// CliLog logs a message to stderr.
func CliLog(message string) { //nolint:revive
	fmt.Fprintln(os.Stderr, message)
}

// MakeDefaultCommands creates a map of default commands for the authmanager.
func MakeDefaultCommands(authClientManager auth.ClientManager) map[string]SubCommand {
	commands := make(map[string]SubCommand)
	commands["create-client"] = &CreateClientCommand{
		clientManager: authClientManager,
	}
	commands["list-clients"] = &ListClientsCommand{
		clientManager: authClientManager,
	}
	commands["get-client"] = &GetClientCommand{
		clientManager: authClientManager,
	}
	commands["delete-client"] = &DeleteClientCommand{
		clientManager: authClientManager,
	}
	return commands
}

// RunWithCommands runs the authmanager with the given commands available.
func RunWithCommands(commands map[string]SubCommand) {
	usage := func() string {
		usage := "usage: authmanager <subcommand>\n"
		usage += "subcommands:"
		for name := range commands {
			usage += fmt.Sprintf(" %s", name)
		}
		return usage
	}

	if len(os.Args) < 2 {
		CliLog(usage())
		os.Exit(1)
	}
	commandName := os.Args[1]
	command, ok := commands[commandName]
	if !ok {
		CliLog(fmt.Sprintf("unknown subcommand: %s", commandName))
		CliLog(usage())
		os.Exit(1)
	}

	if err := command.Run(); err != nil {
		CliLog(fmt.Sprintf("ERROR: %s", err))
		os.Exit(1)
	}
}

// CreateClientCommand is a subcommand for creating a new client.
type CreateClientCommand struct {
	clientManager auth.ClientManager
}

// Run implements the SubCommand interface.
func (c *CreateClientCommand) Run() error {
	cmd := flag.NewFlagSet("create-client", flag.ExitOnError)
	clientID := cmd.String("client-id", "", "client id")
	scope := cmd.String("scope", "", "space separated list of scopes to allow")
	allowIngest := cmd.Bool("allow-ingest", false, "include scopes that allow the client to ingest data")
	secret := cmd.String("client-secret", "", "client secret [generated if not provided]")
	owner := cmd.String("owner", "", "owner of the client")
	clientName := cmd.String("client-name", "", "name of the client")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *clientID == "" {
		return fmt.Errorf("client id is required")
	}

	scopes := []string{}
	if *allowIngest {
		scopes = append(scopes, authutil.ScopeDiodeIngest)
	}
	if *scope != "" {
		scopes = append(scopes, strings.Split(*scope, " ")...)
	}
	if len(scopes) == 0 {
		return fmt.Errorf("one or more scopes are required")
	}
	allScopes := strings.Join(scopes, " ")

	secretProvided := true
	if *secret == "" {
		secretProvided = false
		newSecret, err := auth.GenerateClientSecret()
		if err != nil {
			return fmt.Errorf("failed to generate client secret: %w", err)
		}
		secret = &newSecret
	} else {
		if len(*secret) < 16 {
			return fmt.Errorf("client secret must be at least 16 characters long")
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	created, err := c.clientManager.CreateClient(ctx, auth.ClientInfo{
		ClientID:     *clientID,
		Scope:        allScopes,
		ClientSecret: *secret,
		ClientName:   *clientName,
		Owner:        *owner,
	})
	if err != nil {
		return err
	}
	CliLog("client created successfully.")
	if secretProvided {
		// will only be output if we generated it.
		created.ClientSecret = ""
	}

	bytes, err := json.MarshalIndent(created, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %w", err)
	}

	if !secretProvided {
		CliLog(
			"** NOTE: The client secret is only displayed once and cannot be retrieved later.\n" +
				"  Store credentials in a secure location. If you lose them, you will need to\n" +
				"  destroy and regenerate the client.",
		)
	}

	// output to stdout
	fmt.Println(string(bytes))

	return nil
}

// ListClientsCommand is a subcommand for listing clients.
type ListClientsCommand struct {
	clientManager auth.ClientManager
}

// Run implements the SubCommand interface.
func (c *ListClientsCommand) Run() error {
	cmd := flag.NewFlagSet("list-clients", flag.ExitOnError)
	owner := cmd.String("owner", "", "filter by owner")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	cur := ""
	all := make([]auth.ClientInfo, 0)

	for {
		req := auth.RetrieveClientsRequest{
			PageToken: cur,
		}
		if *owner != "" {
			req.Owner = *owner
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()

		res, err := c.clientManager.RetrieveClients(ctx, req)
		if err != nil {
			return err
		}

		all = append(all, res.Clients...)

		if res.NextPageToken == "" {
			break
		}

		cur = res.NextPageToken
	}

	bytes, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %w", err)
	}
	fmt.Println(string(bytes))

	return nil
}

// GetClientCommand is a subcommand for getting a client.
type GetClientCommand struct {
	clientManager auth.ClientManager
}

// Run implements the SubCommand interface.
func (c *GetClientCommand) Run() error {
	cmd := flag.NewFlagSet("get-client", flag.ExitOnError)
	clientID := cmd.String("client-id", "", "client id")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *clientID == "" {
		return fmt.Errorf("client id is required")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	client, err := c.clientManager.RetrieveClientByID(ctx, *clientID)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %w", err)
	}
	fmt.Println(string(bytes))

	return nil
}

// DeleteClientCommand is a subcommand for deleting a client.
type DeleteClientCommand struct {
	clientManager auth.ClientManager
}

// Run implements the SubCommand interface.
func (c *DeleteClientCommand) Run() error {
	cmd := flag.NewFlagSet("delete-client", flag.ExitOnError)
	clientID := cmd.String("client-id", "", "client id")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *clientID == "" {
		return fmt.Errorf("client id is required")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := c.clientManager.DeleteClientByID(ctx, *clientID)
	if err != nil {
		return err
	}

	CliLog(fmt.Sprintf("client %s deleted successfully", *clientID))
	return nil
}
