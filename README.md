# timespent

A server-based time tracking tool with a focus on getting it done with as little clicks as possible.

## Why?

If you are like me and like tracking your time (say, to make sure you meet daily goals), chances are you do it all **by hand** in an Excel spreadsheet, combined with a separate stopwatch tool.

For a programmer though, it is unacceptable to spend so many unnecessary clicks and keystrokes if you can just automate it all away.

With timespent, you just choose your activity and press "Start". When finished, you press "Stop" and the software does the rest.

## Features

* **Pause/resume function** – no need to spend energy on subtracting interruption times yourself.
* **Daily statistics** – see how many hours you've spent on which activity on a given day.
* **Nightshift mode** – when starting a session after midnight, you get the option to assign it to the statistics of the day before.
* **Cross-device** – the stopwatch keeps running, even if you close the browser and access it from another device.

## How to use

The app is live on [http://217.160.234.15/](http://217.160.234.15/) if you want to use it yourself.

## How to build

1. Clone the repo

2. Fire up a [PostgreSQL](https://www.postgresql.org/docs/current/tutorial.html) server and put the DB connection string into a .env file in the root directory (call the field ```'DB_CONNECTION_STRING'``` like so: ```DB_CONNECTION_STRING="postgresql://joelramilison:@localhost:5432/timespent?sslmode=disable"```

3. Install [Goose](https://github.com/pressly/goose) and run ```goose postgres [DB connection string without the ?sslmode parameter] up``` from the sql/schema directory.

4. Make sure you have Go installed and run ```go build``` in the project folder.
