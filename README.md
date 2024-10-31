What program does: run web server, listen for correct call and write call's param(url param('value') or json body('UUID')) to db.

This program uses Sqllite3 dbms. Required github.com/ncruces/go-sqlite3/driver.

Flags:
* logsDir
* httpPort
* mode

'mode' value may be:
* body(default) - run web server(default port is 3000) and listen for POST with application/json body like ```{"UUID" : "data$12345",    "creationDate" : "2021:08:01 14:22"} ``` URL will be: ```https://<your address>:3000/api```
* param - run web server(default port is 3000) and listen for POST with "value" parameter in URL like 
    ```https://<your address>:3000/api?value=<your param value>```

Workflow is follow:
1. run web server(default port is 3000) and listen for POST according to 'mode' parameter.
    ```https://<your address>:3000/api?value=<your param value>```
2. then if  param is found and it's not empty it's written to the data.db

DB is simple: 
    table 'Data' with columns:
        ID(INTEGER PRIMARY KEY), 
        Value(TEXT NOT NULL UNIQUE), 
        Posted_Date(TEXT),
        Processed(INTEGER(0(failed)/1(succeeded)/NULL(na)))
        Processed_Date(TEXT)
```
CREATE TABLE "Data" (
	"ID"	INTEGER,
	"Value"	TEXT NOT NULL UNIQUE,
	"Posted_Date"	TEXT,
	"Processed"	INTEGER,
	"Processed_Date"	TEXT,
	PRIMARY KEY("ID")
);
```

data_BLANK.db - is just empty DB with stucture described above. 
Rename it to data.db to use with application.

Application params:
    -log-dir: path to logdir; default is 'logs' in program's root
    -port: port of web-srv to listen; default is 3000

