# appland-cli
[![Build Status](https://travis-ci.com/applandinc/appland.svg?token=oNqy5hPadVE4PUAF9ZWk&branch=master)](https://travis-ci.com/applandinc/appland)

## Usage
### Quickstart
```
$ appland login
logging into https://app.land

login: username
password:

logged in.

$ appland stats tmp/appmap/minitest
11655 calls, top 20 methods
  Digest::Instance#digest: 3864 (40 distinct)
  IO#write: 1926 (697 distinct)
  Logger#add: 1846 (249 distinct)
  IO#read: 1558 (7 distinct)
  IO#close: 625 (1 distinct)
  SessionsHelper#current_user: 404 (1 distinct)
  UsersHelper#gravatar_for: 231 (172 distinct)
  SessionsHelper#current_user?: 194 (133 distinct)
  JSON::Ext::Generator::State#generate: 188 (87 distinct)
  OpenSSL::Cipher#final: 136 (1 distinct)
  JSON::Ext::Parser#parse: 122 (1 distinct)
  OpenSSL::Cipher#encrypt: 92 (1 distinct)
  SessionsHelper#logged_in?: 82 (1 distinct)
  User.digest: 58 (25 distinct)
  OpenSSL::Cipher#decrypt: 44 (1 distinct)
  User#feed: 43 (1 distinct)
  ApplicationHelper#full_title: 34 (15 distinct)
  User.new_token: 26 (1 distinct)
  SessionsHelper#log_in: 22 (22 distinct)
  SessionsController#create: 21 (1 distinct
$ appland upload tmp/appmap/minitest/
 100% |████████████████████████████████████████|  [3s:0s]

Success! rails_sample_app_6th_ed has been updated with 67 scenarios.
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

`upload [files, dirs]`
Uploads a list of scenario files or directories to AppLand.

#### stats
Some some statistics about events in scenario files.

`stats [files, dirs]`

## Displaying statistics
The `stats` subcommand will show some simple statistics about events in a collection of
app maps:

```
$ appland stats --help
Show statistics for AppMaps

Usage:
  appland stats [flags]

Flags:
  -f, --files       show statistics for each file
  -h, --help        help for stats
  -j, --json        format results as JSON
  -l, --limit int   limit the number of methods displayed (default 20)
  -p, --params      show distinct parameters for each method
  -v, --verbose     be verbose while processing
```

### Some examples
#### With defaults
Show the top 20 methods. For each method, the total number of call events are shown, as
well as the number of distinct calls (i.e. with different parameters).

```
$ appland stats Application_page_with_a_mapset_restores_the_tab_from_location_hash.appmap.json
229 calls, top 20 methods
  JSON::Ext::Parser#parse: 38 (1 distinct)
  Net::HTTP#request: 32 (8 distinct)
  JSON::Ext::Generator::State#generate: 30 (13 distinct)
  ClassMap::CodeObjectName.parse: 7 (7 distinct)
  Mapset::Show#web_resources: 7 (1 distinct)
  Mapset::Show#data_model: 6 (1 distinct)
  OpenSSL::Cipher#final: 6 (1 distinct)
  WebResources.dehydrate: 5 (2 distinct)
  Configuration#attributes: 4 (1 distinct)
  DataModel.dehydrate: 4 (1 distinct)
  Mapset::Show#recording_method_counts: 4 (1 distinct)
  Search#filter: 4 (2 distinct)
  App::Show#mapsets: 3 (1 distinct)
  Configuration#attributes=: 3 (1 distinct)
  Configuration.find: 3 (3 distinct)
  KeyDataStats.count: 3 (3 distinct)
  OpenSSL::Cipher#decrypt: 3 (1 distinct)
  OpenSSL::Cipher#encrypt: 3 (1 distinct)
  Scenario::SearchActions.search: 3 (3 distinct)
  User.find_by_id!: 3 (1 distinct)
```

#### Statistics for individual files
With `--files`, show statistics for individual files. Adding `--params` will include the
distinct parameters for each method.

```
$ appland stats --files --params Application_page_with_a_mapset_restores_the_tab_from_location_hash.appmap.json
Application_page_with_a_mapset_restores_the_tab_from_location_hash.appmap.json: 229 calls, top 20 methods
  JSON::Ext::Parser#parse: 38 (1 distinct)
   no parameters
  Net::HTTP#request: 32 (8 distinct)
   has parameters
    Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/elements],<nil>,<nil>
    Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/execute/sync],<nil>
    Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/execute/sync],<nil>,<nil>
    Net::HTTP::Get[GET /session/c300eeef64aa2b0dcd284b14cfeca788/element/95946e42-80a6-46a3-ac3b-e3a8c20,<nil>
    Net::HTTP::Get[GET /session/c300eeef64aa2b0dcd284b14cfeca788/element/95946e42-80a6-46a3-ac3b-e3a8c20,<nil>,<nil>
    Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/url],<nil>
    Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/url],<nil>,<nil>
    Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/elements],<nil>
  JSON::Ext::Generator::State#generate: 30 (13 distinct)
   has parameters
    [{"name"=>"controllers", "type"=>"package", "children"=>[{"name"=>"SessionController", "type"=>"clas
    {"HTTP server requests"=>1, "SQL queries"=>1, "Messages"=>1}
    {:using=>"css selector", :value=>"#datamodel.active.show"}
    {:script=>"return ((function(){function d(t,e,n){function r(t){var e=x(t);if(0<e.height&&0<e.width)r
    {:using=>"css selector", :value=>".tab-content .show"}
    {"session_id"=>"d142aeeff345540b576465126fb2d26c", "user_id"=>1, "configuration"=>"{}", "flash"=>{"d
    {"tables"=>["users"], "joins"=>[]}
    {"resource_map"=>{"sessions"=>[{"path"=>"/sessions/new", "methods"=>{"POST"=>[{"event_id"=>1, "scena
    {"packages"=>["HTTP", "SQL"], "class_package"=>{"POST /sessions/new"=>"HTTP", "SQL"=>"SQL"}, "packag
    {"element-6066-11e4-a52e-4f735466cecf"=>"95946e42-80a6-46a3-ac3b-e3a8c2058a82"}
    {:url=>"http://127.0.0.1:49651/applications/1#datamodel"}
    [{"owner_type"=>"mapset", "owner_id"=>1, "recording_method"=>"rspec", "num_scenarios"=>1}]
...
```

#### JSON output
The output can also be formatted as JSON. The elements of the array are sorted by the
number of calls.
```
$ appland stats --json --files --params Application_page_with_a_mapset_restores_the_tab_from_location_hash.appmap.json
Application_page_with_a_mapset_restores_the_tab_from_location_hash.appmap.json: [
  {
    "method": "JSON::Ext::Parser#parse",
    "calls": 38,
    "num_params": 0,
    "param_counts": {
      "": 38
    }
  },
  {
    "method": "Net::HTTP#request",
    "calls": 32,
    "num_params": 2,
    "param_counts": {
      "Net::HTTP::Get[GET /session/c300eeef64aa2b0dcd284b14cfeca788/element/95946e42-80a6-46a3-ac3b-e3a8c20,\u003cnil\u003e": 1,
      "Net::HTTP::Get[GET /session/c300eeef64aa2b0dcd284b14cfeca788/element/95946e42-80a6-46a3-ac3b-e3a8c20,\u003cnil\u003e,\u003cnil\u003e": 1,
      "Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/elements],\u003cnil\u003e": 11,
      "Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/elements],\u003cnil\u003e,\u003cnil\u003e": 11,
      "Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/execute/sync],\u003cnil\u003e": 3,
      "Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/execute/sync],\u003cnil\u003e,\u003cnil\u003e": 3,
      "Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/url],\u003cnil\u003e": 1,
      "Net::HTTP::Post[POST /session/c300eeef64aa2b0dcd284b14cfeca788/url],\u003cnil\u003e,\u003cnil\u003e": 1
    }
  },
  {
    "method": "JSON::Ext::Generator::State#generate",
    "calls": 30,
    "num_params": 1,
    "param_counts": {
      "[{\"name\"=\u003e\"controllers\", \"type\"=\u003e\"package\", \"children\"=\u003e[{\"name\"=\u003e\"SessionController\", \"type\"=\u003e\"clas": 2,
      "[{\"owner_type\"=\u003e\"mapset\", \"owner_id\"=\u003e1, \"recording_method\"=\u003e\"rspec\", \"num_scenarios\"=\u003e1}]": 1,
      "{\"HTTP server requests\"=\u003e1, \"SQL queries\"=\u003e1, \"Messages\"=\u003e1}": 1,
      "{\"element-6066-11e4-a52e-4f735466cecf\"=\u003e\"95946e42-80a6-46a3-ac3b-e3a8c2058a82\"}": 3,
      "{\"package\"=\u003e2, \"class\"=\u003e2, \"function\"=\u003e3}": 1,
      "{\"packages\"=\u003e[\"HTTP\", \"SQL\"], \"class_package\"=\u003e{\"POST /sessions/new\"=\u003e\"HTTP\", \"SQL\"=\u003e\"SQL\"}, \"packag": 1,
      "{\"resource_map\"=\u003e{\"sessions\"=\u003e[{\"path\"=\u003e\"/sessions/new\", \"methods\"=\u003e{\"POST\"=\u003e[{\"event_id\"=\u003e1, \"scena": 1,
      "{\"session_id\"=\u003e\"d142aeeff345540b576465126fb2d26c\", \"user_id\"=\u003e1, \"configuration\"=\u003e\"{}\", \"flash\"=\u003e{\"d": 3,
      "{\"tables\"=\u003e[\"users\"], \"joins\"=\u003e[]}": 2,
      "{:script=\u003e\"return ((function(){function d(t,e,n){function r(t){var e=x(t);if(0\u003ce.height\u0026\u00260\u003ce.width)r": 3,
      "{:url=\u003e\"http://127.0.0.1:49651/applications/1#datamodel\"}": 1,
...
```

## Building
`./bin/build` will build a binary to the `dist` directory. To install, copy the
binary to `/usr/local/bin`.

## Testing
`go test -v ./...` will run all tests. API calls are mocked and do not require a
live AppLand service.

## Releases
Releases are automatically published upon creation of a new tag.
Example:
```bash
$ git tag -a $(cat VERSION) -m "Version $(cat VERSION)"
$ git push origin "$(cat VERSION)"
```
