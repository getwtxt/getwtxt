# getwtxt [![Go Report Card](https://goreportcard.com/badge/github.com/getwtxt/getwtxt)](https://goreportcard.com/report/github.com/getwtxt/getwtxt) [![Build Status](https://travis-ci.com/getwtxt/getwtxt.svg?branch=master)](https://travis-ci.com/getwtxt/getwtxt)

twtxt registry written in Go! 

twtxt is a decentralized microblogging platform "for hackers" based
on text files. The user is "followed" and "mentioned" by referencing the URL to
their `twtxt.txt` (or other text) file and a (not necessarily unique) nickname.
Registries are designed to aggregate several users' statuses into a single location,
facilitating the discovery of new users to follow and allowing the search of statuses
for tags and key words.

Until now, there has primarily been a single registry application available for 
enthusiasts to use to run their own `twtxt` registry. Why not add some diversity
to the landscape?

\[ [Installation](#installation) \] \[ [Configuration](#configuration) \] \[ [Using the Registry](#using-the-registry) \] \[ [Benchmarks](#benchmarks) \] \[ [Notes](#notes) \]

## Features

* Easy to set up and maintain. 
* Uses an in-memory cache to serve requests.
* Pushes to `LevelDB` at a configurable interval for data storage. 
* Run directly facing the internet or behind `Caddy` / `nginx`.

A public instance is currently available:
* [twtxt.tilde.institute](https://twtxt.tilde.institute)

## Installation 

First, fetch the sources using either the `go` tool or using `git` (if you haven't 
set up a `go` development environment and just want to try it out) and jump into
the directory.

```
$ go get github.com/getwtxt/getwtxt
...
$ cd $GOPATH/src/github.com/getwtxt/getwtxt
```

```
$ git clone git://github.com/getwtxt/getwtxt.git
...
$ cd getwtxt
```

Optionally, use the `go` tool to test and benchmark it:

```
$ go test -v -bench . -benchmem
```

Use the `go` tool to build:

```
$ go build -v
```

## Configuration

\[ [Proxying](#proxying) \] \[ [Starting getwtxt](#starting-getwtxt) \]

To configure `getwtxt`, you'll first need to open `getwtxt.yml` in your favorite
editor and modify any values necessary. There are comments in the file explaining
each option. Additionally, you may run `getwtxt` with the `-m` flag for a short
configuration manual.

```
$ ./getwtxt -m | less
```

If you desire, you may additionally modify the template in `assets/tmpl/index.html`
to customize the page users will see when they pull up your registry instance in
a web browser. The values in the configuration file under `Instance:` are used
to replace text `{{.Like This}}` in the template.

### Proxying

Though `getwtxt` will run perfectly fine facing the internet directly, it does not
understand virtual hosts, nor does it use TLS (yet). You'll probably want to proxy it behind
`Caddy` or `nginx` for this reason. 

`Caddy` is ludicrously easy to set up, and automatically handles `TLS` certificates. Here's the config:

```caddyfile
twtxt.example.com 
proxy / example.com:9001
```

If you're using `nginx`, here's a skeleton config to get you started:

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
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_pass http://127.0.0.1:9010;
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

### Starting `getwtxt`

Once you've customized the configuration, start it within a `tmux` session (or `screen` works) and detach.
If you're using a configuration file not in one of the expected locations or with a non-default name, 
start `getwtxt` like this:

```
$ ./getwtxt -c /path/to/configuration/file.yml
```

Otherwise, just:

```
$ ./getwtxt
```

## Using the Registry

The following examples will all apply to using `curl` from a `Linux`, `BSD`, or `macOS` terminal.
All timestamps are in `RFC3339` format, per the twtxt registry specification

### Adding a User
Both nickname and URL are required
```
$ curl -X POST 'https://twtxt.example.com/api/plain/users?url=https://mysite.ext/twtxt.txt&nickname=FooBarringtonJr'

200 OK
```

### Fetch All Statuses
```
$ curl 'https://twtxt.example.com/api/plain/tweets'

foo_barrington  https://foo.barrington.ext/twtxt.txt  2019-03-01T09:31:02.000Z Hey everybody! It's my first status!
...
...

```

### Fetch All Users
Timestamp for when they were added to the registry.

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

### Query by Tag
```
$ curl 'https://twtxt.example.com/api/plain/tags/programming'

foo    https://example.com/twtxt.txt    2019-03-01T09:31:02.000Z    I love #programming!
```

### Query Tweets by Keyword
```
$ curl 'https://twtxt.example.com/api/plain/tweets?q=getwtxt'
foo_barrington    https://example3.com/twtxt.txt    2019-04-30T06:00:09.000Z    I just built getwtxt, time to set it up!

```

## Benchmarks

* [bombardier](https://github.com/codesenberg/bombardier)

```
$ bombardier -c 100 -n 200000 http://localhost:9001/api/plain/tweets

Bombarding http://localhost:9001/api/plain/tweets with 200000 request(s) using 100 connection(s)
 200000 / 200000 [==============================================================================] 100.00% 15100/s 13s

Done!

Statistics        Avg      Stdev        Max
  Reqs/sec     15249.12    3526.87   25047.46
  Latency        6.56ms     2.93ms    64.54ms
  HTTP codes:
    1xx - 0, 2xx - 200000, 3xx - 0, 4xx - 0, 5xx - 0
    others - 0
  Throughput:     7.83MB/s
```

## <a name="notes"></a>Notes

twtxt Information
  * [twtxt.readthedocs.io](https://twtxt.readthedocs.io)

Registry Specification
  * [twtxt.readthedocs.io/.../registry.html](https://twtxt.readthedocs.io/en/latest/user/registry.html)
