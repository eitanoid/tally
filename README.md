# 

CLI example:
```sh
$ tally schema list
> No tally schemas found.

$ tally schema create -n 'Reading Tracker' \
    -m 'Tracking reading habits' \
    -f 'book:which book was read:string:true' \
    -f 'pages:pages read:integer:true' \
    -f 'time:time when reading started:timestamp:true'
> Created schema for 'Reading Tracker' (3Jrf7GMstuu7PrzljZ6hNVidCmf, v1):
> {
>   "tally_id": "3Jrf7GMstuu7PrzljZ6hNVidCmf",
>   "version": 1,
>   "name": "Reading Tracker",
>   "description": "Tracking reading habits",
>   "json_schema": "{\"type\":\"object\",\"properties\":{\"book\":{\"type\":\"string\",\"description\":\"which book was read\"},\"pages\":{\"type\":\"integer\",\"description\":\"pages read\"},\"time\":{\"type\":\"string\",\"description\":\"time when reading started\",\"format\":\"date-time\"}},\"description\":\"Tracking reading habits\",\"required\":[\"book\",\"pages\",\"time\"],\"additionalProperties\":false}",
>   "created_at": "2026-09-26T11:58:58Z"
> }

$ tally schema list
> TALLY ID                      VERSION   NAME              CREATED AT
> --------------------------------------------------------------------
> 3Jrf7GMstuu7PrzljZ6hNVidCmf   v1        Reading Tracker   2026-09-26T11:58:58Z

$ tally entry list -t 3Jrf7GMstuu7PrzljZ6hNVidCmf
> No entries found for tally '3Jrf7GMstuu7PrzljZ6hNVidCmf'.

$ tally entry log -t 3Jrf7GMstuu7PrzljZ6hNVidCmf -d "{\"book\":\"The Stranger\", \"pages\": 20, \"time\":\"$(date -Iseconds)\"}"
> Logged entry 3JrfPNjVVGFmo7htV8Y8WssmwFp for tally '3Jrf7GMstuu7PrzljZ6hNVidCmf' (v1) at 2026-09-26 12:01:22

$ tally entry list -t 3Jrf7GMstuu7PrzljZ6hNVidCmf
> ENTRY ID                      VERSION   CREATED AT             DATA
> -------------------------------------------------------------------
> 3JrfPNjVVGFmo7htV8Y8WssmwFp   v1        2026-09-26T12:01:22Z   {"book":"The Stranger", "pages": 20, "time":"2026-09-26T13:01:22+01:00"}

$ tally entry log -t 3Jrf7GMstuu7PrzljZ6hNVidCmf -d "{\"book\":\"The Fall\", \"pages\": 23, \"time\":\"$(date -Iseconds)\"}"
> Logged entry 3JrfSUesJ1TLrVM0gR86dpDDr1Q for tally '3Jrf7GMstuu7PrzljZ6hNVidCmf' (v1) at 2026-09-26 12:01:47

$ tally entry log -t 3Jrf7GMstuu7PrzljZ6hNVidCmf -d "{\"book\":\"The Fall\", \"pages\": 23, \"time\":\"$(date -Iseconds)\", \"review\":\"good\"}"
> Error: failed to log entry: entry domain creation failed: failed to validate entry data: entry validation failed: validating root: unexpected additional properties ["review"]

$ tally entry list -t 3Jrf7GMstuu7PrzljZ6hNVidCmf
> ENTRY ID                      VERSION   CREATED AT             DATA
> -------------------------------------------------------------------
> 3JrfSUesJ1TLrVM0gR86dpDDr1Q   v1        2026-09-26T12:01:47Z   {"book":"The Fall", "pages": 23, "time":"2026-09-26T13:01:47+01:00"}
> 3JrfPNjVVGFmo7htV8Y8WssmwFp   v1        2026-09-26T12:01:22Z   {"book":"The Stranger", "pages": 20, "time":"2026-09-26T13:01:22+01:00"}
```
Generic habit tracking app for Android (in the future)

Features:
- Create a new habit with a row schema (eg. Medicine: name string, dose-mg int, time timestamp)
- Add a new entry to the habit following the schema
- Browse your habit entries

Technical details:

- golang app deployed to android (with potentially a different frontend later?)
- go migrate and a sqlite3 driver for go avoiding CGO

2 tables in one sqlite database
- `entries` where each habit is stored and contains a json schema
- `habits` where each entry is associated to an entry in schemas and acocmpanies a body object as well as a creation timestamp
