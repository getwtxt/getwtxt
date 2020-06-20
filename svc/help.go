/*
Copyright (c) 2019 Ben Morrison (gbmor)

This file is part of Getwtxt.

Getwtxt is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Getwtxt is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Getwtxt.  If not, see <https://www.gnu.org/licenses/>.
*/

package svc // import "git.sr.ht/~gbmor/getwtxt/svc"

import "fmt"

func titleScreen() {
	fmt.Printf(`

                       _            _        _
             __ _  ___| |___      _| |___  _| |_
            / _  |/ _ \ __\ \ /\ / / __\ \/ / __|
           | (_| |  __/ |_ \ V  V /| |_ >  <| |_
            \__, |\___|\__| \_/\_/  \__/_/\_\\__|
            |___/
                       version ` + Vers + `
                   git.sr.ht/~gbmor/getwtxt
                          GPL  v3

`)
}

func helpScreen() {
	fmt.Printf(`
                        getwtxt Help


                 :: Command Line Options ::

    Command line options are used to explicitly override defaults,
 or what has been specified in the configuration file.

    -h [--help]      Print this help screen.
    -m [--manual]    Print the manual.
    -v [--version]   Print the version information and quit.
    -c [--config]    Path to an alternate configuration file
                       to use. May be relative or absolute.
    -a [--assets]    Path to the assets directory, containing
                       style.css and tmpl/index.html
    -d [--db]        Path getwtxt should use for the database.
    -t [--dbtype]    Type of database to use.
                       Options: leveldb (default)
                                sqlite

`)
}

func manualScreen() {
	fmt.Printf(`
                       :: Sections ::

    >> Configuration File
        Covers syntax and location of default configuration,
        passing a specific configuration file to getwtxt,
        and acceptable formats for configuration files.

    >> Customizing the Landing Page
        Covers the location of the landing page template,
        format of the template, and optional preprocessor
        tags available to use when creating a new landing
        page template.

    >> Interacting With the Registry
        Explains all API endpoints, their parameters,
        and expected output.


                  :: Configuration File ::

    The default configuration file is in YAML format, chosen for
 its clarity and its support of commenting (unlike JSON). It may
 be placed in any of the following locations by default:

    The same directory as the getwtxt executable
    /usr/local/getwtxt/
    /etc/
    /usr/local/etc/

    The paths are searched in that order. The first configuration
 file found is used by getwtxt, while the locations further down
 are ignored.

    Multiple configuration files may be used, however, with the
 '-c' command line flag. The path passed to getwtxt via '-c' may
 be relative or absolute. For example, both of the following are
 allowed:

    ./getwtxt -c myconfig.json
    ./getwtxt -c /etc/ExtraConfigsDir/mysecondconfig.toml

 The supported configuration types are:
    YAML, TOML, JSON, HCL

    The configuration file contains several options used to
 customize your instance of getwtxt. None are required, they will
 simply use their default value unless otherwise specified.

    BehindProxy: Informs getwtxt whether it is behind a
        reverse proxy, such as nginx or Caddy. If set to
        false, getwtxt will use host matching for
        incoming requests. The host matched is the URL
        suboption of Instance in the config file.
        Default: true

    ListenPort: Defines the port getwtxt should bind to.
        Default: 9001

    UseTLS: Boolean value that lets getwtxt know if it
        should use TLS for incoming connections.
        Default: false

    TLSCert: Absolute path to the certificate file used
        for TLS connections.
        Default: /etc/ssl/getwtxt.pem

    TLSKey: Absolute path to the private TLS key file
        used for TLS connections.
        Default: /etc/ssl/private/getwtxt.pem

    DatabaseType: The type of back-end getwtxt should use
        to store registry data. The available types of
        databases are: leveldb
                       sqlite
        Default: leveldb

    DatabasePath: The location of the LevelDB structure
        used by getwtxt to back up registry data. This
        can be a relative or absolute path.
        Default: getwtxt.db

    AssetsDirectory: This is the directory where getwtxt
        can find style.css and tmpl/index.html -- the
        stylesheet for the landing page and the landing
        page template, respectively.
        Default: assets

    StdoutLogging: Boolean used to determine whether
        getwtxt should send logging output to stdout.
        This is useful for debugging, but you should
        probably save your logs once your instance
        is running.
        Default: false

    MessageLog: The location of getwtxt's error and
        other messages log file. This, like DatabasePath,
        can be relative or absolute.
        Default: logs/message.log

    RequestLog: The location of getwtxt's request log
        file. The path can be relative or absolute.
        Default: logs/request.log

    DatabasePushInterval: The interval on which getwtxt
        will push registry data from the in-memory cache
        to the on-disk LevelDB database. The following
        time suffixes may be used:
            s, m, h
        Default: 5m

    StatusFetchInterval: The interval on which getwtxt
        will crawl all users' twtxt files to retrieve
        new statuses. The same time suffixes as
        DatabasePushInterval may be used.
        Default: 1h

    Instance: Signifies the start of instance-specific
        meta information. The following are used only
        for the summary and use information displayed
        by the default web page for getwtxt. If desired,
        the assets/tmpl/index.html file may be
        customized. Keep in mind that in YAML, the
        following options must be preceded by two spaces
        so that they are interpreted as sub-options.

    SiteName: The name of your getwtxt instance.
        Default: getwtxt

    URL: The publicly-accessible URL of your
        getwtxt instance.
        Default: https://twtxt.example.com

    OwnerName: Your name.
        Default: Anonymous Microblogger

    Email: Your email address.
        Default: nobody@knows

    Description: A short description of your getwtxt
        instance or your site. As this likely includes
        whitespace, it should be in double-quotes.
        This can include XHTML or HTML line breaks if
        desired:
            <br />
            <br>
        Default: "A fast, resilient twtxt registry
            server written in Go!"


             :: Customizing the Landing Page ::

    If you like, feel free to customize the landing page
 template provided at

        assets/tmpl/index.html

    It must be standard HTML or XHTML. There are a few special
 tags available to use that will be replaced with specific values
 when the template is parsed by getwtxt.

    Values are derived from the "Instance" section of the
 configuration file, except for the version of getwtxt used. The
 following will be in the form of:

    {{.TemplateTag}} What it will be replaced with when
        the template is processed and the landing page is
        served to a visitor.

    The surrounding double braces and prefixed period are required
 if you choose to use these tags in your modified landing page. The
 tags themselves are not required; access to them is provided simply
 for convenience.

    {{.Vers}} The version of getwtxt used. Does not include
        the preceding 'v'. Ex: 0.2.0

    {{.Name}} The name of the instance.

    {{.Owner}} The instance owner's name.

    {{.Mail}} Email address used for contacting the instance
        owner if the need arises.

    {{.Desc}} Short description placed in the configuration
        file. This is why HTML tags are allowed.

    {{.URL}} The publicly-accessible URL of your instance. In
        the default landing page, example API calls are shown
        using this URL for the convenience of the user. This
        is also used as the matched host when the "BehindProxy"
        value in the configuration file is set to false.


              :: Interacting with the Registry ::

    The registry API is rather simple, and can be interacted with
 via the command line using cURL. Example output of the calls will
 not be provided.

    Pseudo line-breaks will be represented with a backslash.
 Examples with line-breaks are not syntactically correct and will
 be rejected by cURL. Please concatenate the example calls without
 the backslash. This is only present to maintain consistent
 formatting for this manual text.

    Ex:
        /api/plain/users\
        ?q=FOO
    Should be:
        /api/plain/users?q=FOO

    All queries (every call except adding users) accept the
 ?page=N parameter, where N > 0. The output is provided in groups
 of 20 results. For example, indexed at 1, ?page=2 (or &page=2 if
 it is not the first parameter) appended to any query will return
 results 21 through 40. If the page requested will exceed the
 bounds of the query output, the last 20 query results are returned.

 Adding a user:
    curl -X POST 'http://localhost:9001/api/plain/users\
        ?url=https://example.org/twtxt.txt&nickname=somebody'

 Retrieve user list:
    curl 'http://localhost:9001/api/plain/users'

 Retrieve all statuses:
    curl 'http://localhost:9001/api/plain/tweets'

 Retrieve all statuses with mentions:
    curl 'http://localhost:9001/api/plain/mentions'

 Retrieve all statuses with tags:
    curl 'http://localhost:9001/api/plain/tags'

 Query for users by keyword:
    curl 'http://localhost:9001/api/plain/users?q=FOO'

 Query for users by URL:
    curl 'http://localhost:9001/api/plain/users\
        ?url=https://gbmor.dev/twtxt.txt'

 Query for statuses by substring:
    curl 'http://localhost:9001/api/plain/tweets\
        ?q=SUBSTRING'

 Query for statuses mentioning a user:
    curl 'http://localhost:9001/api/plain/mentions\
        ?url=https://gbmor.dev/twtxt.txt'

 Query for statuses with a given tag:
    curl 'http://localhost:9001/api/plain/tags/myTagHere'

`)
}
