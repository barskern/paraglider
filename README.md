# paragliding

[![Build Status](https://travis-ci.com/barskern/paragliding.svg?branch=master)](https://travis-ci.com/barskern/paragliding)
[![Go Report Card](https://goreportcard.com/badge/github.com/barskern/paragliding)](https://goreportcard.com/report/github.com/barskern/paragliding)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/barskern/paragliding)
[![Release](https://img.shields.io/github/release/barskern/paragliding.svg)](https://github.com/barskern/paragliding/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/barskern/paragliding/badge.svg?branch=master)](https://coveralls.io/github/barskern/paragliding?branch=master)


# About

An online service that will allow users to browse information about IGC files. IGC is an international file format for soaring track files that are used by paragliders and gliders.

The service will store IGC files metadata in a NoSQL Database (persistent storage). The system will generate events and it will monitor for new events happening from the outside services.

# API-endpoints

## GET /paragliding/api

Returns metadata about the service.

## POST /paragliding/api/track

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


## GET /paragliding/api/track

Returns all the ids of all registered tracks.

```
[<id1>, <id2>, ...]
```

## GET /paragliding/api/track/`<id>`

Returns metadata about a specific track. `<id>` is a valid track id which was returned on insertion using `POST`.

```
{
"H_date": <date from File Header, H-record>,
"pilot": <pilot>,
"glider": <glider>,
"glider_id": <glider_id>,
"track_length": <calculated total track length>,
"track_src_url": <the original URL used to upload the track, ie. the URL used with POST>
}
```

## GET /paragliding/api/track/`<id>`/`<field>`

Possible `<field>`-values:

* `pilot`
* `glider`
* `glider_id`
* `track_length`
* `H_date`
* `track_src_url`
