# appland-cli

## Usage
### Quickstart
```
$ appland login
logging into https://app.land

login: username
password:

logged in.

$ appland upload --org myorg tmp/*.json
 100% |████████████████████████████████████████|  [0s:0s]

Success! Application has been updated with 1 scenarios.
https://app.land/applications/5?mapset=13
```

### Running in CI/CD
Configure your CI/CD tool to provide the following environment variables at
runtime. By providing these environment variables, `appland` can authenticate
without any persistent configuration.
- `APPLAND_API_KEY`: Generate a new API key from your [account page](https://app.land/user) to populate this value.
- `APPLAND_URL`: Typically this will always be set to `https://app.land`


### Commands
#### authentication
Authentication and API key management.

`appland login`  
This will prompt you for a login and password. Your password will not be echoed.

`appland logout`  
Logs the current user out of AppLand and revokes the API key in use.

#### contexts
AppLand has the ability to support a number of configuration contexts. In most
cases, you won't need additional contexts. Upon first run, a `default` context
is created, pointing to [app.land](https://app.land). Subcommands which issue
API calls to an AppLand service (such as `login` and `upload`) will use this
context for configuration options and authentication.

`context add [name] [url]`  
Create a new context.

`context current`  
Display the current context.

`context list`  
Show all available contexts.

`context use [name]`
Select a context as the current context. This is set to a default context upon
first run.

#### upload
Upload a mapset of scenario files.

`upload --org <organization> [files]`  
Uploads an array of scenario files to AppLand.

## Building
`./bin/build` will build a binary to the `dist` directory. To install, copy the
binary to `/usr/local/bin`.

## Testing
`go test -v ./...` will run all tests. API calls are mocked and do not require a
live AppLand service.
