# paragliding

[![Build Status](https://travis-ci.com/barskern/paragliding.svg?branch=master)](https://travis-ci.com/barskern/paragliding)
[![Go Report Card](https://goreportcard.com/badge/github.com/barskern/paragliding)](https://goreportcard.com/report/github.com/barskern/paragliding)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/barskern/paragliding)
[![Release](https://img.shields.io/github/release/barskern/paragliding.svg)](https://github.com/barskern/paragliding/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/barskern/paragliding/badge.svg?branch=master)](https://coveralls.io/github/barskern/paragliding?branch=master)


# About

An online service that will allow users to browse information about IGC files. IGC is an international file format for soaring track files that are used by paragliders and gliders.

The service will store IGC files metadata in a NoSQL Database (persistent storage). The system will generate events, which can be subscribed to using webhooks, and it will monitor for new events happening from the outside services.

# API-endpoints

## `GET /paragliding/api`

Returns metadata about the service formatted as a `json` struct.

## `POST /paragliding/api/track`

Register a track. A single track can only be registered **once**.

### Request

```json
{
  "url": "<url>"
}
```

`<url>` represents a normal URL, that would work in a browser, eg: `http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc`.

### Response

```json
{
  "id": "<id>"
}
```

The returned `<id>` will be a unique identifier for the posted track.


## `GET /paragliding/api/track`

Returns all the ids of all registered tracks.

```json
[<id1>, <id2>, ...]
```

## `GET /paragliding/api/track/<id>`

Returns metadata about a specific track. `<id>` is a valid track id which was returned on insertion using `POST`.

```json
{
"H_date": <date from File Header, H-record>,
"pilot": <pilot>,
"glider": <glider>,
"glider_id": <glider_id>,
"track_length": <calculated total track length>,
"track_src_url": <the original URL used to upload the track, ie. the URL used with POST>
}
```

## `GET /paragliding/api/track/<id>/<field>`

Possible `<field>`-values:

* `pilot`
* `glider`
* `glider_id`
* `track_length`
* `H_date`
* `track_src_url`

The response will be formatted as plain text.

## `GET /api/ticker/latest`

Returns the `timestamp` (formatted as specified in RFC3339) of the last added track as plain text.

## `GET /api/ticker/`

Returns a report of the oldest tracks added.

```json
{
"t_latest": <latest added timestamp>,
"t_start": <the first timestamp of the added track>, this will be the oldest track recorded
"t_stop": <the last timestamp of the added track>, this might equal to t_latest if there are no more tracks left
"tracks": [<id1>, <id2>, ...],
"processing": <time in ms of how long it took to process the request>
}
```

## `GET /api/ticker/<timestamp>`

Returns a report of the added tracks after a certain timestamp (formatted as specified in RFC3339).

```json
{
"t_latest": <latest added timestamp of the entire collection>,
"t_start": <the first timestamp of the added track>, this will be higher than the parameter provided in the query
"t_stop": <the last timestamp of the added track>, this might equal to t_latest if there are no more tracks left
"tracks": [<id1>, <id2>, ...],
"processing": <time in ms of how long it took to process the request>
}
```
