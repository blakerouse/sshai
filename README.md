# SSH AI
SSH AI is an MCP server that runs locally on your host that provides the ability to add and remove
hosts that you can SSH into and perform commands.

## Tools

- Add host
  - Adds a host
- Remove host
  - Removes a host
- Get OS Info
  - Shows the OS information of the hosts
- Update OS Info
  - Updates the cached OS information of the hosts
- Perform Command
  - Performs the command on the provided hosts

## Limitations

- Only supports Linux
- Only supports username/password SSH
- Ignores host keys


## How to Setup

Checkout the repository:

```shell
$ git clone https://github.com/blakerouse/sshai
```

Build the binary:

```shell
$ go build .
```

Update the MCP configuration for Claude Desktop:

```
{
  "mcpServers": {
    "ssh": {
      "command": "<PATH_TO_BUILT_SSHAI_BINARY>",
      "args": ["--storage", "<PATH_TO_STORE_HOSTS>", "--openai", "<OPENAI_API_KEY>"]
    }
  }
}
```

Restart Claude Desktop

## How to Use

First add a host that you want to be able to interact with:

`add host with name <name> connecting with ssh://<USER>:<PASS>@<IP>`

You can then list the hosts available with:

`list my hosts`

You can add another host with same command again (multiple hosts will perform the same
command on the provided hosts):

`add host with name <name> connecting with ssh://<USER>:<PASS>@<IP>`

Ask for the OS information for all added hosts:

`show OS information for all hosts`

Ask for it to check if any of the hosts need to be updated:

`do any of my hosts need to be updated`

Ask it to upgrade a specific host (or all, it will do it at the same time):

`upgrade host <name>`

## Considerations

### How would you test the agent, what are the different failure scenarios, and what tools or methods would you use to manage them?

Other than unit testing, overall evaluations would need to be performed with different models to
evaluate the commands and actions. To do this infrastructure needs to be setup with CI to create
VM's in specific states that will allow the AI to use the MCP server to control. Then specific
prompts would be passed through the AI to validate that the expected actions are performed.

### What are the security considerations of your agent and what are potential ways to defend against them?

At the moment this is only setup to run locally on a machine and is not a remote MCP server. If this
was a remote MCP server their is a host of issues related to the fact that RBAC needs to be added
to the hosts that are added to the storage. This could be managed directly in the MCP server or that
could be handled by an external service. It would need to be clear that the machine that the user
is interacting with is one that they are allowed to work on.

### What are the performance concerns for your agent, how will it scale in throughout, response times, and supported tasks?

At the moment this is doing parallel tasks for all added hosts. This is great to streamline the
same action across multiple hosts, but as the host list grows this can be a problem. Adding gaurd
rails to the MCP and limiting the number of hosts that can be worked on at one time would be nice
to have. The tasks it is performing can be long running, being able to stream the work as it is
occurring would be great to add as well. This would allow feedback to the user while work is being
performed instead of it just sitting and spinning.

### What caveats should be documented or gotchas?

At the moment this only supports Linux. This would need to do more work to ensure that it works
both with macOS and Windows.

### If you had more time, what would you do differently, and how would you expand the functionality?

With more time I would set it up to automatically connect to gcloud or aws to fetch the hosts. This
would streamline the onboarding process of adding hosts and keep it in sync with the current state
of VM's in the organization.

I would expand this to add support for macOS and Windows. Long running tasks are also an issue
I would see if I could make those types of tasks stream the work as its occurring back to the user
and allow them to view the work as it occurs.
