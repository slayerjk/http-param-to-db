What program does: run web server, listen for correct call and write call's param(url param('value') or json body('UUID')) to db.

This program uses Sqllite3 dbms. Required github.com/ncruces/go-sqlite3/driver.

Flags:

* log-dir - path to logs dir; default is relative to exe - 'logs_http-param-to-db'
* port - custom port for http server; default is 3000
* mode - body/param; default is 'body'; read below
* param-name - param to seek in request body/param and insert into db 
* body-condition - optional condition to be met in 'body' request, format is 'key:value'; read below

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
Normally, must be as such:
```
{
    "YOUR-PARAM":"***",
    "some-value":***, 
    "maybe-another-value":***
}
```

There is additional flag for condition to be met in json body in format "key:value".
If condition is not "", then accept POST only if condition is found in reques body.
This 'key-value' pair also must not be nested!

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

<h3>mailing</h3>
There is also in 'data' dir mailing_BLANK.json. 
Rename it to 'mailing.json' and fill with your values.
For now only for unauthorized SMTP:25.
Mail for errors and success for processed queries.

mailing.json
```
{
    "host": "my-smtp@example.com",
    "port": 25,
    "from_addr": "my-app@example.com",
    "to_addr_errors": [
        "admin-1@example.com",
        "admin-2@example.com"
    ],
    "to_addr_reports": [
        "user-1@example.com",
        "user-2@example.com"
    ]
}
```

<h3>Examples</h3>

# Run in mode 'param': waits for POST to URL: http://<your server ip>/api?UUID=<value> and writes <value> to DB(only UNIQUE)
> ./http-param-to-db.exe -mode param 

# Run in mode 'param': waits for POST to URL: http://<your server host>/api?my-value=<value> and writes <value> of 'my-value' param to DB(only UNIQUE)
> ./http-param-to-db.exe -mode param -param-name my-value

# Run in mode 'body'(by default): waits for POST to URL: http://<your server host>/api with not nested 'UUID' param(by default) and 'type' = 'waitingLines' (key:value) in JSON body. Writes 'UUID' value to DB(only UNIQUE)
> ./http-param-to-db.exe -body-condition type:waitingLines
