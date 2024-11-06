What program does: run web server, listen for correct call and write call's param(url param('value') or json body('UUID')) to db.

This program uses Sqllite3 dbms. Required github.com/ncruces/go-sqlite3/driver.

Flags:
* log-dir
* keep-logs
* port
* mode
* param-name

<h2>'mode' flag values</h2>

<h3>'body'</h3>
* body(default) - run web server(default port is 3000) and listen for POST with application/json body.
    Param to take in json is by default = 'UUID'. Can be changed with flag 'param-name' and in must exist in json body.
    URL will be: ```https://<your address>:3000/api```

<b>Program will try to search in root of Json, so your param must not be nested.</b>

It's hard to predict how will json body look like, so consider your param is not nested.

Example of json body('UUID' will be found and, for example, 'delegate' will not):
```
{
    "UUID": "",
    "title": "",
    "type": "",
    "creationDate": {
      "delegate": "",
      "isDateTime": true
    },
    "message": {
      "lang": "",
      "text": ""
    }
  }
```

<h3>'param'</h3>
* param - run web server(default port is 3000) and listen for POST with "value" parameter in URL like 
    ```https://<your address>:3000/api?value=<your param value>```

Param to parse in request is by default = 'UUID'. Can be changed with flag 'param-name' and in must exist in callback.json.

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