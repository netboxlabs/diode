# Tests
This directory contains integrations tests that can be run against the Diode Plugin


Here's what you'll need to do in order to run these tests:
- Start the Docker container for Diode Plugin
- Create the user and his token
- Configure the test settings
- Run behave


## Start the Docker container for Diode Plugin

To run the tests, you must have the diode plugin directory and execute the following command in diode/diode-server folder:

```bash
make docker-compose-up
```

## Create the user and his token

To create the user, execute:

```bash
docker exec -it diode-netbox-1 /opt/netbox/netbox/manage.py createsuperuser
```
Fill the username and password for the superuser as requested.

With this user, you can access the Netbox at http://0.0.0.0:8000/ and using the menu Admin -> API Token, you can create the token for this user.

## Test settings
Create the test config file from the template: `cp config.ini.tpl config.ini`.

Then fill in the correct values:

- **user_token**:
  - Mandatory!
  - string
  - user token created in the previous step

- **api_root_path**:
  - Mandatory!
  - string
  - netbox API URL, e.g. http://0.0.0.0:8000/api

- **api_key**:
  - Mandatory!
  - string
  - INGESTION_API_KEY


## Run behave using parallel process

You can use [behavex](https://github.com/hrcorval/behavex) to run the scenarios using multiprocess by simply run:

Examples:

> behavex -t @\<TAG\> --parallel-processes=2 --parallel-schema=scenario

> behavex -t @\<TAG\> --parallel-processes=2 --parallel-schema=feature

Running smoke tests:

> behavex -t=@smoke --parallel-processes=2 --parallel-scheme=scenario


## Test execution reports
[behavex](https://github.com/hrcorval/behavex) provides a friendly HTML test execution report that contains information related to test scenarios, execution status, execution evidence and metrics. A filters bar is also provided to filter scenarios by name, tag or status.

It should be available at the following path:

<output_folder>/report.html

## Clean your environment

The tests clean up the environment after running, you do not need any manual intervention to clean up the environment.
