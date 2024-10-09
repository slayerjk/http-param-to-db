# simple web(req param) to db
Run web server, listen for correct call and write call's param to db

This program uses Sqllite dbms.

DB is simple: 
    table requests with columns:
        ID(pk), Name(text, NOT NULL), Is_Processed(0(failed)/1(succeeded)/NULL(na))
