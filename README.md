# getwtxt [![Build Status](https://travis-ci.com/getwtxt/getwtxt.svg?branch=master)](https://travis-ci.com/getwtxt/getwtxt)

twtxt registry written in Go! 

* Easy to set up and maintain. 
* Uses an in-memory cache to serve requests.
* Pushes to `LevelDB` at a configurable interval for data storage. 
* Run directly facing the internet or behind `Caddy` / `nginx`.

## Development Progress

`ETA` 31 May 2019


* [x] Types and Config
* [x] HTTP Routing
* [x] Registry Manipulation ([getwtxt/registry](https://github.com/getwtxt/registry))
* [x] Request Handling
* [x] Cache Management
* [x] Persistent Data Storage
* [ ] Refactor / Test / Debug
* [ ] Documentation

## Notes

* twtxt
  * [twtxt.readthedocs.io](https://twtxt.readthedocs.io)
