# igcinfo

[![Build Status](https://travis-ci.com/barskern/igcinfo.svg?branch=master)](https://travis-ci.com/barskern/igcinfo)
[![Go Report Card](https://goreportcard.com/badge/github.com/barskern/igcinfo)](https://goreportcard.com/report/github.com/barskern/igcinfo)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/barskern/igcinfo)
[![Release](https://img.shields.io/github/release/barskern/igcinfo.svg)](https://github.com/barskern/igcinfo/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/barskern/igcinfo/badge.svg?branch=master)](https://coveralls.io/github/barskern/igcinfo?branch=master)


# About

An online service that will allow users to browse information about IGC files. IGC is an international file format for soaring track files that are used by paragliders and gliders. The program will not store anything in a persistent storage. I.e. no information will be stored on the server side on a disk or database. Instead, it will store submitted tracks in memory. Subsequent API calls will allow the user to browse and inspect stored IGC files.

# API-endpoints

## GET /igcinfo/api

Returns metadata about the service.

```
{
  "uptime": <uptime>
  "info": "Service for IGC tracks."
  "version": "v1"
}
```

`<uptime>` is the current uptime of the service formatted according to [Duration format as specified by ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations).

## POST /igcinfo/api/igc

Register a track

### Request

```
{
  "url": "<url>"
}
```

`<url>` represents a normal URL, that would work in a browser, eg: `http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc`.

### Response

```
{
  "id": "<id>"
}
```

The returned `<id>` will be a unique identifier for the posted track.


## GET /igcinfo/api/igc

Returns all the ids of all registered tracks.

```
[<id1>, <id2>, ...]
```

## GET /igcinfo/api/igc/`<id>`

Returns metadata about a specific track. `<id>` is a valid track id which was returned on insertion using `POST`.

```
{
"H_date": <date from File Header, H-record>,
"pilot": <pilot>,
"glider": <glider>,
"glider_id": <glider_id>,
"track_length": <calculated total track length>
}
```

## GET /igcinfo/api/igc/`<id>`/`<field>`

Possible `<field>`-values:

* `pilot`
* `glider`
* `glider_id`
* `track_length`
* `H_date`
