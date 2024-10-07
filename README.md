# Jiwa - A CLI interface for Jira that hopefully doesn't suck (on your soul)

The basic idea is that you get a fully pipeable interface to Jira, nothing less nothing more.
Functionality will be somewhat limited especially around searching currently because that's complicated
and I need to think about it a bit more. Overall most of the stuff should be explained in `jiwa <command> --help`, but this
very new code so if something is wrong please reach out, either via issue or mail: `catouc@philipp.boeschen.me`.

An example on what you can do:

```shell
cat ticket-file | jiwa create -i - | jiwa reassign $user | jiwa label on-call urgent | jiwa mv "in progress"
```

By default, if you call `jiwa create`, you can control the behaviour of it with `--in or -i`, it looks up your `$EDITOR` variable and provides a similar interface to
`git commit`, as in the first line is what will be the ticket title. The description follows separated by a new line:

```
Summary line of my ticket

Description that can be quite long
and span multiple lines.
```

# Installation

```
go install github.com/catouc/jiwa/cmd/jiwa@latest
```

# Configuration

Jiwa currently uses a configuration file under `$HOME/.config/jiwa/config.json` that needs to be filled with:

```json
{
  "baseURL": "https://catouc.atlassian.net",
  "username": "atlassian@philipp.boeschen.me",
  "password": "<pass>"
}
```

You can alternatively set `JIWA_USERNAME` and `JIWA_PASSWORD` in your environment and that will have the same effect.
For token based authentication you need to set `token` or `JIWA_TOKEN` instead and can omit the password variable.

If you instance has weird prefixes in the URLs you can use `endpointPrefix` like:

```json
{
  "endpointPrefix": "/jira"
}
```

(until I get around to it that leading `/` is very important!).

# Developing

My own test instance is at https://catouc.atlassian.net/jira/software/projects/JIWA/boards/1
The username is my atlassian account mail and I can generate an API token under https://id.atlassian.com/manage-profile/security/api-tokens
The token needs to then be in `JIWA_PASSWORD` because OAuth1 is used where you jam that into the password field?
