# Tests
This directory contains integrations tests that can be run against the Diode Plugin


Here's what you'll need to do in order to run these tests:
- Start docker containers stack (diode and NetBox)
- Check the users and their tokens
- Configure the test settings
- Run behave


## Start the Docker container for Netbox with Diode Plugin

To run the tests, you must have the diode plugin directory, and execute the following commands in the **diode-server** folder.

```bash
pip install netboxlabs-diode-netbox-plugin
```

After that, you can start the docker container by running the following command:

```bash
make docker-compose-up
```

## Users and tokens

The command above will create all users necessary to run the tests.

Using the Admin user, you can access the Netbox at http://0.0.0.0:8000/.

- username: admin
- password: admin

To check the tokens of the users, navigate to the "Admin" menu and select "API Token". This will display a list of all the tokens associated with the users.


Please, pay attention to the token for user "INGESTION", it will be used in the next section.

## Test settings
Create the test config file from the template: `cp config.ini.tpl config.ini`.

Then fill in the correct values:

- **user_token**:
  - Mandatory!
  - string
  - **ADMIN** token created in the previous step

- **api_root_path**:
  - Mandatory!
  - string
  - netbox API URL, e.g. http://0.0.0.0:8000/api

- **api_key**:
  - Mandatory!
  - string
  - **INGESTION** user token created in the previous step


## Run behave using parallel process

You can use [behavex](https://github.com/hrcorval/behavex) to run the scenarios using multiprocess by simply run:

Examples:

> behavex -t @\<TAG\> --parallel-processes=2 --parallel-schema=feature

> behavex -t @\<TAG\> --parallel-processes=2 --parallel-schema=feature

Running smoke tests:

> behavex -t=@smoke --parallel-processes=2 --parallel-scheme=feature


## Test execution reports
[behavex](https://github.com/hrcorval/behavex) provides a friendly HTML test execution report that contains information related to test scenarios, execution status, execution evidence and metrics. A filters bar is also provided to filter scenarios by name, tag or status.

It should be available at the following path:

<output_folder>/report.html

## Clean your environment

After running the tests, clean up your environment by running the command:

> behavex -t=@cleanup --parallel-processes=2 --parallel-scheme=feature