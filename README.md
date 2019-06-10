&nbsp;[![Go Report Card](https://goreportcard.com/badge/github.com/getwtxt/getwtxt)](https://goreportcard.com/report/github.com/getwtxt/getwtxt) 
&nbsp;[![Code Climate maintainability](https://img.shields.io/codeclimate/maintainability-percentage/getwtxt/getwtxt.svg)](https://codeclimate.com/github/getwtxt/getwtxt)
&nbsp;[![Code Climate issues](https://img.shields.io/codeclimate/issues/getwtxt/getwtxt.svg)](https://codeclimate.com/github/getwtxt/getwtxt/issues)
&nbsp;[![Code Climate technical debt](https://img.shields.io/codeclimate/tech-debt/getwtxt/getwtxt.svg)](https://codeclimate.com/github/getwtxt/getwtxt/trends/technical_debt)
# getwtxt &nbsp;[![Build Status](https://travis-ci.com/getwtxt/getwtxt.svg?branch=master)](https://travis-ci.com/getwtxt/getwtxt) 

twtxt registry written in Go! 

[twtxt](https://github.com/buckket/twtxt) is a decentralized microblogging platform 
"for hackers" based on text files. The user is "followed" and "mentioned" by referencing 
the URL to their `twtxt.txt` file and a nickname.
Registries are designed to aggregate several users' statuses into a single location,
facilitating the discovery of new users to follow and allowing the search of statuses
for tags and key words.

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
\[ [Installation](#installation) \] 
&nbsp; &nbsp; \[ [Configuration](#configuration) \] 
&nbsp; &nbsp; \[ [Using the Registry](#using-the-registry) \] 
&nbsp; &nbsp; \[ [Benchmarks](#benchmarks) \] 
&nbsp; &nbsp; \[ [Other Documentation](#other-documentation) \] 
&nbsp; &nbsp; \[ [Notes](#notes) \]

## Features 
&nbsp;[![GitHub release](https://img.shields.io/github/release/getwtxt/getwtxt.svg)](https://github.com/getwtxt/getwtxt/releases/latest)

* Easy to set up and maintain 
* Uses an in-memory cache to serve requests
* Pushes to a database at a configurable interval for persistent storage
  * `leveldb (default)` 
  * `sqlite3`
* More database support is in development
* Run directly facing the internet or behind `Caddy` / `nginx`

A public instance is currently available:
* [twtxt.tilde.institute](https://twtxt.tilde.institute)

## Installation 

While I've included macOS builds in Travis CI, I have only
personally tested getwtxt on Linux, specifically:
* `Debian 9, 10/Testing, Sid`
* `Ubuntu Server 18.04LTS, 18.10, 19.04`

Build dependencies are minimal, and only include:
* `gnu make`
* `go >= 1.11`

`git` is not required if you download the sources via the 
[`Releases`](https://github.com/getwtxt/getwtxt/releases) tab

Now, on with the directions. First, fetch the sources using `git`
and jump into the directory.

```
$ git clone git://github.com/getwtxt/getwtxt.git
...
$ cd getwtxt
```

Optionally, use the `go` tool to test and benchmark the files in `svc`.
If you choose to run the tests, be sure to return to the main directory afterwards.

```
$ cd svc && go test -v -bench . -benchmem
...
...
PASS

$ cd ..
```

Use `make` to initiate the build and install process. 
```
$ make
...
$ sudo make install
```

## Configuration

\[ [Proxying](#proxying) \] &nbsp; \[ [Starting getwtxt](#starting-getwtxt) \]

To configure getwtxt, you'll first need to open `/usr/local/getwtxt/getwtxt.yml` 
in your favorite editor and modify any values necessary. There are comments in the 
file explaining each option.

If you desire, you may additionally modify the template in 
`/usr/local/getwtxt/assets/tmpl/index.html` to customize the page users will see 
when they pull up your registry instance in a web browser. The values in the 
configuration file under `Instance:` are used to replace text `{{.Like This}}` in 
the template.

### Proxying

Though getwtxt will run perfectly fine facing the internet directly, it does not
understand virtual hosts, nor does it use TLS (yet). You'll probably want to proxy it behind
`Caddy` or `nginx` for this reason. 

`Caddy` is ludicrously easy to set up, and automatically handles `TLS` certificates. Here's the config:

```caddyfile
twtxt.example.com 
proxy / example.com:9001
```

If you're using `nginx`, here's a skeleton config to get you started. Don't forget to change 
the 5 instances of `twtxt.example.com` to the (sub)domain you'll be using to access the registry, 
generate SSL/TLS certificates using `letsencrypt`, and change the port in `proxy_pass` to whichever 
port you specified when modifying the configuration file. Currently, it's set to the default port `9001`

```nginx
server {
    server_name twtxt.example.com;
    listen [::]:443 ssl http2;
    listen 0.0.0.0:443 ssl http2;
    ssl_certificate /etc/letsencrypt/live/twtxt.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/twtxt.example.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
    location / {
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_pass http://127.0.0.1:9001;
    }
}
server {
    if ($host = twtxt.example.com) {
        return 301 https://$host$request_uri;
    }
    listen 80;
    server_name twtxt.example.com;
    return 404;
}
```

### Starting getwtxt

Once you have everything configured to your needs, use `systemctl` to enable it
to run on system boot, then start the service.

```
$ sudo systemctl enable getwtxt
...
$ sudo systemctl start getwtxt
```

## Using the Registry

The following examples will all apply to using `curl` from a `Linux`, `BSD`, or `macOS` terminal.
All timestamps are in `RFC3339` format, per the twtxt registry specification. Additionally, all
queries support the `?page=N` parameter, where `N` is a positive integer, that will retrieve page
`N` of results in groups of twenty.

The example API calls can also be found on the landing page of any getwtxt instance, assuming
the admin has not customized the landing page.
* [twtxt.tilde.institute](https://twtxt.tilde.institute)

### Adding a User
Both nickname and URL are required
```
$ curl -X POST 'https://twtxt.example.com/api/plain/users?url=https://mysite.ext/twtxt.txt&nickname=FooJr'

200 OK
```

### Get All Tweets
```
$ curl 'https://twtxt.example.com/api/plain/tweets'

foo_barrington  https://foo.bar.ext/twtxt.txt  2019-03-01T09:31:02.000Z Hey! It's my first status!
...
...
```

### Query Tweets by Keyword
```
$ curl 'https://twtxt.example.com/api/plain/tweets?q=getwtxt'

foo_barrington    https://example3.com/twtxt.txt    2019-04-30T06:00:09.000Z    I just installed getwtxt!
```

### Get All Users
Timestamp reflects when the user was added to the registry.

```
$ curl 'https://twtxt.example.com/api/plain/users'

foo_barrington      https://foo.barrington.ext/twtxt.txt  2017-01-01T09:17:02.000Z
foo_barrington_jr   https://example.com/twtxt.txt         2019-03-01T09:31:02.000Z
...
...
```

### Query Users
Can use either keyword or URL.

```
$ curl 'https://twtxt.example.com/api/plain/users?url=https://example.com/twtxt.txt'

foo               https://example.com/twtxt.txt     2019-05-09T08:42:23.000Z


$ curl 'https://twtxt.example.com/api/plain/users?q=foo'

foo               https://example.com/twtxt.txt     2019-05-09T08:42:23.000Z
foobar            https://example2.com/twtxt.txt    2019-03-14T19:23:00.000Z
foo_barrington    https://example3.com/twtxt.txt    2019-05-01T15:59:39.000Z
```

### Get all tweets with mentions
Mentions are placed within a status using the format `@<nickname http://url/twtxt.txt>`
```
$ curl 'https://twtxt.tilde.institute/api/plain/mentions'

foo               https://example.com/twtxt.txt     2019-02-28T11:06:44.000Z    @<foo_barrington https://example3.com/twtxt.txt> Hey!! Are you still working on that project?
bar               https://mxmmplm.com/twtxt.txt     2019-02-27T11:06:44.000Z    @<foobar https://example2.com/twtxt.txt> How's your day going, bud?
foo_barrington    https://example3.com/twtxt.txt    2019-02-26T11:06:44.000Z    @<foo https://example.com/twtxt.txt> Did you eat my lunch?
```

### Query tweets by mention URL
```
$ curl 'https://twtxt.tilde.institute/api/plain/mentions?url=https://foobarrington.co.uk/twtxt.txt'

foo    https://example.com/twtxt.txt    2019-02-26T11:06:44.000Z    @<foo_barrington https://foobarrington.co.uk/twtxt.txt> Hey!! Are you still working on that project?e
```

### Get all Tags
```
$ curl 'https://twtxt.example.com/api/plain/tags'

foo    https://example.com/twtxt.txt    2019-03-01T09:33:04.000Z    No, seriously, I need #help
foo    https://example.com/twtxt.txt    2019-03-01T09:32:12.000Z    Seriously, I love #programming!
foo    https://example.com/twtxt.txt    2019-03-01T09:31:02.000Z    I love #programming!
```

### Query by Tag
```
$ curl 'https://twtxt.example.com/api/plain/tags/programming'

foo    https://example.com/twtxt.txt    2019-03-01T09:31:02.000Z    I love #programming!
```

## Benchmarks

* [bombardier](https://github.com/codesenberg/bombardier)

```
$ bombardier -c 100 -n 200000 http://localhost:9001/api/plain/tweets

Bombarding http://localhost:9001/api/plain/tweets with 200000 request(s) using 100 connection(s)
 200000 / 200000 [=============================================================] 100.00% 19961/s 10s

Done!

Statistics        Avg      Stdev        Max
  Reqs/sec     20006.58    2408.55   26054.73
  Latency        5.00ms     3.58ms    62.99ms
  HTTP codes:
    1xx - 0, 2xx - 200000, 3xx - 0, 4xx - 0, 5xx - 0
    others - 0
  Throughput:    39.27MB/s	
```

## Other Documentation

In addition to what is provided here, additional information, particularly regarding the configuration
file, may be found by running getwtxt with the `-m` or `--manual` flags. You will likely want to pipe the output
to `less` as it is quite long.

```
$ ./getwtxt -m | less

$ ./getwtxt --manual | less
```

If you need to remove getwtxt from your system, navigate to the source directory
you acquired using `git` during the installation process and run the appropriate
`make` hook:

```
$ sudo make uninstall
```

## Notes

twtxt Information: [`twtxt.readthedocs.io`](https://twtxt.readthedocs.io)

twtxt Client Repo: [`github.com/buckket/twtxt`](https://github.com/buckket/twtxt)

Registry Specification: [`twtxt.readthedocs.io/en/latest/user/registry.html`](https://twtxt.readthedocs.io/en/latest/user/registry.html)

Special thanks to [`github.com/kognise/water.css`](https://github.com/kognise/water.css) for open-sourcing a pleasant, easy-to-use, importable stylesheet
