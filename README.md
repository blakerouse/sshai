# SSH AI
SSH AI is an MCP server that runs locally on your host that provides the ability to SSH into host
and perform actions.

## Tools

- Check for Updates
  - Checks if any updates are available on the host.

## Limitations

- Only supports Ubuntu with APT
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
      "args": ["--openai", "<OPENAI_API_KEY>"]
    }
  }
}
```

Restart Claude Desktop

## How to Use

Ask the following question to Claude and watch it work. Be sure the connection information
includes username, password and that the machine is an Ubuntu machine that allows SSH over
password (by default most do not).

`SSH into ssh://<USER>:<PASS>@<IP> and see if it is up to date`
