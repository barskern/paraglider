# igcinfo

[![Build Status](https://travis-ci.com/barskern/igcinfo.svg?branch=master)](https://travis-ci.com/barskern/igcinfo)
[![Go Report Card](https://goreportcard.com/badge/github.com/barskern/igcinfo)](https://goreportcard.com/report/github.com/barskern/igcinfo)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/barskern/igcinfo)
[![Release](https://img.shields.io/github/release/barskern/igcinfo.svg)](https://github.com/barskern/igcinfo/releases/latest)

# About

Develop an online service that will allow users to browse information about IGC files. IGC is an international file format for soaring track files that are used by paragliders and gliders. The program will not store anything in a persistent storage. I.e. no information will be stored on the server side on a disk or database. Instead, it will store submitted tracks in memory. Subsequent API calls will allow the user to browse and inspect stored IGC files.

For the development of the IGC processing, you will use an open source IGC library for Go: goigc

The system must be deployed on either Heroku or Google App Engine, and the Go source code must be available for inspection by the teaching staff (read-only access is sufficient).

# API-endpoints

- [X] `GET /api`
- [X] `GET /api/igc`
- [X] `POST /api/igc`
- [X] `GET /api/igc/<id>`
- [X] `GET /api/igc/<id>/<field>`
