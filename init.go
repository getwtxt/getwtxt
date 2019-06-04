package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getwtxt/registry"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

const getwtxt = "0.3.0"

var (
	flagVersion  *bool   = pflag.BoolP("version", "v", false, "Display version information, then exit.")
	flagHelp     *bool   = pflag.BoolP("help", "h", false, "Display the quick-help screen.")
	flagMan      *bool   = pflag.BoolP("manual", "m", false, "Display the configuration manual.")
	flagConfFile *string = pflag.StringP("config", "c", "", "The name/path of the configuration file you wish to use.")
)

var confObj = &Configuration{}

// signals to close the log file
var closeLog = make(chan bool, 1)

// used to transmit database pointer after
// initialization
var dbChan = make(chan *leveldb.DB, 1)

var tmpls *template.Template

var twtxtCache = registry.NewIndex()

var remoteRegistries = &RemoteRegistries{}

var staticCache = &struct {
	index    []byte
	indexMod time.Time
	css      []byte
	cssMod   time.Time
}{
	index:    nil,
	indexMod: time.Time{},
	css:      nil,
	cssMod:   time.Time{},
}

func initGetwtxt() {
	checkFlags()
	titleScreen()
	initConfig()
	initLogging()
	tmpls = initTemplates()
	initDatabase()
	watchForInterrupt()
}

func checkFlags() {
	pflag.Parse()
	if *flagVersion {
		titleScreen()
		os.Exit(0)
	}
	if *flagHelp {
		titleScreen()
		helpScreen()
		os.Exit(0)
	}
	if *flagMan {
		titleScreen()
		helpScreen()
		manualScreen()
		os.Exit(0)
	}
}

func initConfig() {

	if *flagConfFile == "" {
		viper.SetConfigName("getwtxt")
		viper.SetConfigType("yml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/usr/local/getwtxt")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("/usr/local/etc")

	} else {
		path, file := filepath.Split(*flagConfFile)
		if path == "" {
			path = "."
		}
		if file == "" {
			file = *flagConfFile
		}
		filename := strings.Split(file, ".")
		viper.SetConfigName(filename[0])
		viper.SetConfigType(filename[1])
		viper.AddConfigPath(path)
	}

	log.Printf("Loading configuration ...\n")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("%v\n", err.Error())
		log.Printf("Using defaults ...\n")
	} else {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("Config file change detected. Reloading...\n")
			rebindConfig()
		})
	}

	viper.SetDefault("ListenPort", 9001)
	viper.SetDefault("LogFile", "getwtxt.log")
	viper.SetDefault("DatabasePath", "getwtxt.db")
	viper.SetDefault("StdoutLogging", false)
	viper.SetDefault("ReCacheInterval", "1h")
	viper.SetDefault("DatabasePushInterval", "5m")

	viper.SetDefault("Instance.SiteName", "getwtxt")
	viper.SetDefault("Instance.OwnerName", "Anonymous Microblogger")
	viper.SetDefault("Instance.Email", "nobody@knows")
	viper.SetDefault("Instance.URL", "https://twtxt.example.com")
	viper.SetDefault("Instance.Description", "A fast, resilient twtxt registry server written in Go!")

	confObj.Mu.Lock()

	confObj.Port = viper.GetInt("ListenPort")
	confObj.LogFile = viper.GetString("LogFile")

	confObj.DBPath = viper.GetString("DatabasePath")
	log.Printf("Using database: %v\n", confObj.DBPath)

	confObj.StdoutLogging = viper.GetBool("StdoutLogging")
	if confObj.StdoutLogging {
		log.Printf("Logging to stdout\n")
	} else {
		log.Printf("Logging to %v\n", confObj.LogFile)
	}

	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	log.Printf("User status fetch interval: %v\n", confObj.CacheInterval)

	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")
	log.Printf("Database push interval: %v\n", confObj.DBInterval)

	confObj.LastCache = time.Now()
	confObj.LastPush = time.Now()
	confObj.Version = getwtxt

	confObj.Instance.Vers = getwtxt
	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")

	confObj.Mu.Unlock()

}

func initLogging() {

	confObj.Mu.RLock()

	if confObj.StdoutLogging {
		log.SetOutput(os.Stdout)

	} else {

		logfile, err := os.OpenFile(confObj.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Printf("Could not open log file: %v\n", err.Error())
		}

		// Listen for the signal to close the log file
		// in a separate thread. Passing it as an argument
		// to prevent race conditions when the config is
		// reloaded.
		go func(logfile *os.File) {

			<-closeLog
			log.Printf("Closing log file ...\n")

			err = logfile.Close()
			if err != nil {
				log.Printf("Couldn't close log file: %v\n", err.Error())
			}
		}(logfile)

		log.SetOutput(logfile)
	}

	confObj.Mu.RUnlock()
}

func rebindConfig() {

	confObj.Mu.RLock()
	if !confObj.StdoutLogging {
		closeLog <- true
	}
	confObj.Mu.RUnlock()

	confObj.Mu.Lock()

	confObj.LogFile = viper.GetString("LogFile")
	confObj.DBPath = viper.GetString("DatabasePath")
	confObj.StdoutLogging = viper.GetBool("StdoutLogging")
	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")

	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")

	confObj.Mu.Unlock()

	initLogging()
}

func initTemplates() *template.Template {
	return template.Must(template.ParseFiles("assets/tmpl/index.html"))
}

// Pull DB data into cache, if available.
func initDatabase() {
	confObj.Mu.RLock()
	db, err := leveldb.OpenFile(confObj.DBPath, nil)
	confObj.Mu.RUnlock()
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	dbChan <- db

	pullDatabase()
	go cacheAndPush()
}

// Watch for SIGINT aka ^C
// Close the log file then exit
func watchForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {

			log.Printf("\n\nCaught %v. Cleaning up ...\n", sigint)
			confObj.Mu.RLock()

			log.Printf("Closing database connection to %v...\n", confObj.DBPath)
			db := <-dbChan
			if err := db.Close(); err != nil {
				log.Printf("%v\n", err.Error())
			}

			if !confObj.StdoutLogging {
				closeLog <- true
			}

			confObj.Mu.RUnlock()
			close(dbChan)
			close(closeLog)

			// Let everything catch up
			time.Sleep(100 * time.Millisecond)
			os.Exit(0)
		}
	}()
}

func titleScreen() {
	fmt.Printf(`
    
                       _            _        _
             __ _  ___| |___      _| |___  _| |_
            / _  |/ _ \ __\ \ /\ / / __\ \/ / __|
           | (_| |  __/ |_ \ V  V /| |_ >  <| |_
            \__, |\___|\__| \_/\_/  \__/_/\_\\__|
            |___/
                       version ` + getwtxt + `
                 github.com/getwtxt/getwtxt
                          GPL  v3  

`)
}

func helpScreen() {
	fmt.Printf(`
                        getwtxt Help


                 :: Command Line Options ::

Command Line Options:
    -h [--help]      Print this help screen.
    -m [--manual]    Print the manual.
    -v [--version]   Print the version information and quit.
    -c [--config]    Path to an alternate configuration file
                       to use. May be relative or absolute.

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

    ListenPort: Defines the port getwtxt should bind to.
        Default: 9001

    DatabasePath: The location of the LevelDB structure
        used by getwtxt to back up registry data. This
        can be a relative or absolute path.
        Default: getwtxt.db

    StdoutLogging: Boolean used to determine whether
        getwtxt should send logging output to stdout.
        This is useful for debugging, but you should
        probably save your logs once your instance 
        is running.
        Default: false

    LogFile: The location of getwtxt's log file. This,
        like DatabasePath, can be relative or absolute.
        Default: getwtxt.log

    DatabasePushInterval: The interval on which getwtxt
        will push registry data from the in-memory cache
        to the on-disk LevelDB database. The following
        time suffixes may be used:
            ns, us, ms, s, m, h
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
        using this URL for the convenience of the user.


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
        ?url=https://gbmor.dev/twtxt.txt&nickname=gbmor'

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
