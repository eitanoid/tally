# 

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
