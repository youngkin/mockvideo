# SQL

The 'sql' directory contains the sql needed to create the mockvideo project's database and tables. It also includes a sample test data file.

Files `createTables.sh` and `createTestData.sh` may need to be modified for the local database environment. Specifically:

* `-uadmin` references a user named `admin`. This may need to be changed.
* `-pXXXXX` references the password for user `admin`. It  needs to be set to whatever the password for the user specified in `-u`.
* `-h10.0.0.100` references the host address for the `mysql` server. This setting must reflect the correct address for the `mysql` server which may be different than this.
